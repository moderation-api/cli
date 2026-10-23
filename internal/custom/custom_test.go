package custom

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectStack(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  stack
	}{
		{
			name:  "next with pnpm",
			files: map[string]string{"package.json": `{"dependencies":{"next":"15"}}`, "pnpm-lock.yaml": "", "tsconfig.json": "{}"},
			want:  stack{Language: "typescript", Framework: "Next.js", PackageManager: "pnpm", SDKPackage: nodeSDK, InstallCommand: "pnpm add " + nodeSDK, EnvFile: ".env.local"},
		},
		{
			name:  "express with the sdk",
			files: map[string]string{"package.json": `{"dependencies":{"express":"4","@moderation-api/sdk":"2"}}`},
			want:  stack{Language: "javascript", Framework: "Express", PackageManager: "npm", SDKPackage: nodeSDK, SDKInstalled: true, InstallCommand: "npm install " + nodeSDK, EnvFile: ".env"},
		},
		{
			name:  "django with uv",
			files: map[string]string{"pyproject.toml": "dependencies = [\"Django>=5\"]", "uv.lock": ""},
			want:  stack{Language: "python", Framework: "Django", PackageManager: "uv", SDKPackage: pythonSDK, InstallCommand: "uv add " + pythonSDK, EnvFile: ".env"},
		},
		{
			name:  "rails",
			files: map[string]string{"Gemfile": "gem \"rails\"\ngem \"moderation_api\""},
			want:  stack{Language: "ruby", Framework: "Rails", PackageManager: "bundler", SDKPackage: "moderation_api", SDKInstalled: true, InstallCommand: "bundle add moderation_api", EnvFile: ".env"},
		},
		{
			name:  "laravel",
			files: map[string]string{"composer.json": `{"require":{"laravel/framework":"^11"}}`},
			want:  stack{Language: "php", Framework: "Laravel", PackageManager: "composer", SDKPackage: phpSDK, InstallCommand: "composer require " + phpSDK, EnvFile: ".env"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tc.files {
				write(t, dir, name, content)
			}
			got := detectStack(dir)
			if got == nil {
				t.Fatal("no stack detected")
			}
			got.Note = ""
			if *got != tc.want {
				t.Fatalf("got %+v\nwant %+v", *got, tc.want)
			}
		})
	}

	if s := detectStack(t.TempDir()); s != nil {
		t.Fatalf("empty dir: got %+v, want nil", s)
	}
	dir := t.TempDir()
	write(t, dir, "go.mod", "module x")
	if s := detectStack(dir); s == nil || s.Language != "go" || s.SDKPackage != "" || s.Note == "" {
		t.Fatalf("go: got %+v", s)
	}
}

func TestSetEnvValuesReplacesAndAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	write(t, filepath.Dir(path), ".env", "# app\nDATABASE_URL=postgres://x\nexport MODAPI_SECRET_KEY=old\n")
	err := setEnvValues(path, map[string]string{projectKeyEnv: "sk_new", webhookSecretEnv: "whsec"}, []string{projectKeyEnv, webhookSecretEnv})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	want := "# app\nDATABASE_URL=postgres://x\nexport MODAPI_SECRET_KEY=sk_new\nMODAPI_WEBHOOK_SECRET=whsec\n"
	if string(data) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", data, want)
	}
	if got := readEnvValue(path, projectKeyEnv); got != "sk_new" {
		t.Fatalf("readEnvValue = %q", got)
	}
}

func TestSetEnvValuesCreatesPrivateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.local")
	if err := setEnvValues(path, map[string]string{projectKeyEnv: "sk_1"}, []string{projectKeyEnv}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode = %o, want 600", perm)
	}
}

func TestReadEnvValueStripsQuotes(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".env", `MODAPI_SECRET_KEY="sk_quoted"`+"\n")
	if got := readEnvValue(filepath.Join(dir, ".env"), projectKeyEnv); got != "sk_quoted" {
		t.Fatalf("got %q", got)
	}
}

func TestPKCEChallengeIsS256OfVerifier(t *testing.T) {
	p, err := newPKCE()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Verifier) != 43 {
		t.Fatalf("verifier length = %d, want 43", len(p.Verifier))
	}
	sum := sha256.Sum256([]byte(p.Verifier))
	if p.Challenge != base64.RawURLEncoding.EncodeToString(sum[:]) {
		t.Fatal("challenge is not base64url(sha256(verifier))")
	}
}

// fakeClerk is a minimal authorization server for the login flows.
type fakeClerk struct {
	*httptest.Server
	devicePolls atomic.Int32
	pending     int32
	lastToken   url.Values
}

func newFakeClerk(t *testing.T, pending int32) *fakeClerk {
	f := &fakeClerk{pending: pending}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                        f.URL,
			"authorization_endpoint":        f.URL + "/oauth/authorize",
			"token_endpoint":                f.URL + "/oauth/token",
			"device_authorization_endpoint": f.URL + "/oauth/device_authorization",
		})
	})
	mux.HandleFunc("/oauth/device_authorization", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("client_id") != "cli" {
			http.Error(w, `{"error":"invalid_client"}`, http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code": "dev-1", "user_code": "ABCD-EFGH",
			"verification_uri": f.URL + "/device", "expires_in": 600, "interval": 5,
		})
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.lastToken = r.Form
		w.Header().Set("Content-Type", "application/json")
		switch r.Form.Get("grant_type") {
		case "urn:ietf:params:oauth:grant-type:device_code":
			if f.devicePolls.Add(1) <= f.pending {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"authorization_pending"}`)
				return
			}
		case "authorization_code":
			if r.Form.Get("code") != "code-1" || r.Form.Get("code_verifier") == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
				return
			}
		}
		_, _ = io.WriteString(w, `{"access_token":"at-1","token_type":"Bearer","expires_in":300}`)
	})
	f.Server = httptest.NewServer(mux)
	t.Cleanup(f.Close)
	return f
}

func TestDeviceLoginPollsUntilApproved(t *testing.T) {
	clerk := newFakeClerk(t, 2)
	server, err := discover(context.Background(), clerk.Client(), clerk.URL)
	if err != nil {
		t.Fatal(err)
	}
	var prompt bytes.Buffer
	var slept []time.Duration
	tokens, err := deviceLogin(context.Background(), deviceLoginOptions{
		Server: server, ClientID: "cli", Scopes: loginScopes, Prompt: &prompt, HTTP: clerk.Client(),
		Sleep: func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if tokens.AccessToken != "at-1" {
		t.Fatalf("access token = %q", tokens.AccessToken)
	}
	if len(slept) != 3 || slept[0] != 5*time.Second {
		t.Fatalf("polled %d times with %v", len(slept), slept)
	}
	if !strings.Contains(prompt.String(), "ABCD-EFGH") {
		t.Fatalf("prompt does not show the user code:\n%s", prompt.String())
	}
}

func TestBrowserLoginExchangesCodeWithVerifier(t *testing.T) {
	clerk := newFakeClerk(t, 0)
	server, err := discover(context.Background(), clerk.Client(), clerk.URL)
	if err != nil {
		t.Fatal(err)
	}
	// The "browser" follows the authorize URL's redirect with a code.
	open := func(raw string) error {
		authURL, err := url.Parse(raw)
		if err != nil {
			return err
		}
		q := authURL.Query()
		if q.Get("code_challenge_method") != "S256" || q.Get("resource") != "https://example.com/cli" {
			t.Errorf("authorize URL is missing PKCE or resource: %s", raw)
		}
		redirect := q.Get("redirect_uri") + "?code=code-1&state=" + url.QueryEscape(q.Get("state"))
		go func() {
			res, err := http.Get(redirect)
			if err == nil {
				res.Body.Close()
			}
		}()
		return nil
	}
	tokens, err := browserLogin(context.Background(), browserLoginOptions{
		Server: server, ClientID: "cli", Resource: "https://example.com/cli", Scopes: loginScopes,
		Open: open, Prompt: io.Discard, Wait: 10 * time.Second, HTTP: clerk.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if tokens.AccessToken != "at-1" {
		t.Fatalf("access token = %q", tokens.AccessToken)
	}
	if !strings.HasPrefix(clerk.lastToken.Get("redirect_uri"), "http://127.0.0.1:") {
		t.Fatalf("token request redirect_uri = %q", clerk.lastToken.Get("redirect_uri"))
	}
}

func TestCallbackRejectsWrongState(t *testing.T) {
	results := make(chan callbackResult, 1)
	rec := httptest.NewRecorder()
	callbackHandler("expected", results).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/callback?code=x&state=other", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	select {
	case r := <-results:
		t.Fatalf("a request with the wrong state ended the sign-in: %+v", r)
	default:
	}

	// The real redirect still gets through afterwards.
	rec = httptest.NewRecorder()
	callbackHandler("expected", results).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/callback?code=c1&state=expected", nil))
	if r := <-results; r.err != nil || r.code != "c1" {
		t.Fatalf("got %+v", r)
	}
}

func TestDiscoverRejectsWrongIssuer(t *testing.T) {
	clerk := newFakeClerk(t, 0)
	if _, err := discover(context.Background(), clerk.Client(), "https://other.example.com"); err == nil {
		t.Fatal("metadata for another issuer was accepted")
	}
}

func TestRequireHTTPS(t *testing.T) {
	for raw, ok := range map[string]bool{
		"https://dash.moderationapi.com": true,
		"http://127.0.0.1:8080":          true,
		"http://localhost:3000":          true,
		"http://dash.moderationapi.com":  false,
	} {
		if err := requireHTTPS(raw); (err == nil) != ok {
			t.Errorf("requireHTTPS(%q) = %v", raw, err)
		}
	}
}

func TestEnvValueEdgeCases(t *testing.T) {
	for raw, want := range map[string]string{
		`abc # note`:   "abc",
		`"a b" # note`: "a b",
		`'x'`:          "x",
		`"esc\"q"`:     `esc"q`,
	} {
		if got := parseEnvValue(raw); got != want {
			t.Errorf("parseEnvValue(%q) = %q, want %q", raw, got, want)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	write(t, dir, ".env", "export MODAPI_SECRET_KEY=old\n")
	if err := setEnvValues(path, map[string]string{projectKeyEnv: "has space"}, []string{projectKeyEnv}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "export MODAPI_SECRET_KEY=\"has space\"\n" {
		t.Fatalf("got %q", data)
	}
	if got := readEnvValue(path, projectKeyEnv); got != "has space" {
		t.Fatalf("round trip = %q", got)
	}
}

func TestExchangeForOrganizationKey(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/v1/cli/key":
			_, _ = io.WriteString(w, `{"apiKey":"sk_org","organizationId":"org_1","organizationName":"Acme"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got, err := exchangeForOrganizationKey(context.Background(), srv.Client(), srv.URL+"/v1/cli/key", "at-1", "laptop")
	if err != nil {
		t.Fatal(err)
	}
	if got.APIKey != "sk_org" || got.OrganizationName != "Acme" || gotAuth != "Bearer at-1" {
		t.Fatalf("got %+v, auth %q", got, gotAuth)
	}

	_, err = exchangeForOrganizationKey(context.Background(), srv.Client(), srv.URL+"/missing", "at-1", "laptop")
	if !errors.Is(err, errExchangeNotAvailable) {
		t.Fatalf("404 should mean not available, got %v", err)
	}
}

func TestKeyExchangeURLIsTheTokenAudience(t *testing.T) {
	t.Setenv("MODAPI_KEY_EXCHANGE_URL", "")
	t.Setenv("MODAPI_OAUTH_RESOURCE", "")
	cfg := loadSettings()
	if cfg.KeyExchangeURL != "https://api.moderationapi.com/v1/cli/key" || cfg.Resource != cfg.KeyExchangeURL {
		t.Fatalf("got exchange %q, resource %q", cfg.KeyExchangeURL, cfg.Resource)
	}

	t.Setenv("MODAPI_KEY_EXCHANGE_URL", "https://staging.example.com/v1/cli/key")
	if cfg := loadSettings(); cfg.Resource != "https://staging.example.com/v1/cli/key" {
		t.Fatalf("resource %q does not follow the exchange URL", cfg.Resource)
	}
}
