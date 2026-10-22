package custom_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/moderation-api/cli/internal/cli"
	"github.com/moderation-api/cli/internal/clierrors"
	"github.com/moderation-api/cli/internal/config"
	"github.com/moderation-api/cli/internal/output"
)

// fakeAPI answers the endpoints init and doctor call, and records requests.
type fakeAPI struct {
	*httptest.Server
	mu       sync.Mutex
	requests []recorded
}

type recorded struct {
	Method, Path, Auth string
	Body               map[string]any
}

func newFakeAPI(t *testing.T) *fakeAPI {
	f := &fakeAPI{}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.requests = append(f.requests, recorded{r.Method, r.URL.Path, r.Header.Get("Authorization"), body})
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		auth := r.Header.Get("Authorization")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects":
			if auth != "Bearer sk_org_test" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"message":"bad key","code":"unauthorized"}`)
				return
			}
			_, _ = io.WriteString(w, `{"id":"proj_1","name":"`+body["name"].(string)+`","description":null,"domain":"acme.com","context":"A marketplace","globalFlaggingMode":"ACTIVE","enableRequestLogging":true,"casebookEnabled":true,"createdAt":"2026-10-22T00:00:00Z","updatedAt":"2026-10-22T00:00:00Z","secretKey":"sk_proj_new"}`)
		case r.URL.Path == "/webhook-secret":
			_, _ = io.WriteString(w, `{"secret":"whsec_new"}`)
		case r.URL.Path == "/account":
			if auth != "Bearer sk_proj_new" && auth != "Bearer sk_proj_existing" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"message":"bad key","code":"unauthorized"}`)
				return
			}
			_, _ = io.WriteString(w, `{"id":"acc_1","paid_plan_name":"Pro","text_api_quota":1000,"remaining_quota":900,"current_project":{"id":"proj_1","name":"market"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.Close)
	return f
}

func (f *fakeAPI) paths() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, r := range f.requests {
		out = append(out, r.Method+" "+r.Path)
	}
	return out
}

// run executes modapi with args in a clean config home and returns stdout,
// stderr and the exit code.
func run(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	config.Reset()
	output.ResetAgentMode()
	output.ResetPreparsedRendering()
	root, err := cli.NewRootCommand()
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetIn(strings.NewReader(""))
	err = cli.ExecuteRoot(context.Background(), root, args)
	code := 0
	if err != nil {
		code = clierrors.ExitCode(err)
		if !output.IsRendered(err) {
			stderr.WriteString(err.Error())
		}
	}
	return stdout.String(), stderr.String(), code
}

func setup(t *testing.T) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MODAPI_NO_KEYRING", "true")
	for _, name := range []string{"MODAPI_SECRET_KEY", "MODAPI_WEBHOOK_SECRET", "MODAPI_ORGANIZATION_KEY", "CLAUDECODE", "CURSOR_AGENT", "CODEX"} {
		t.Setenv(name, "")
		_ = os.Unsetenv(name)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"@acme/market","dependencies":{"next":"15"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("git"); err == nil {
		_ = exec.Command("git", "-C", dir, "init", "-q").Run()
	}
	return dir
}

func TestInitCreatesProjectAndWritesEnv(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")

	stdout, stderr, code := run(t, "init", "--dir", dir, "--server-url", api.URL,
		"--domain", "acme.com", "--context", "A marketplace", "--brief", "No scams", "-o", "json")
	if code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if result["status"] != "created" || result["env_file"] != ".env.local" {
		t.Fatalf("unexpected result: %s", stdout)
	}
	if strings.Contains(stdout, "sk_proj_new") || strings.Contains(stdout, "whsec_new") {
		t.Fatal("init printed a secret")
	}

	env, err := os.ReadFile(filepath.Join(dir, ".env.local"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "MODAPI_SECRET_KEY=sk_proj_new\nMODAPI_WEBHOOK_SECRET=whsec_new\n"; string(env) != want {
		t.Fatalf(".env.local = %q, want %q", env, want)
	}

	api.mu.Lock()
	create := api.requests[0]
	api.mu.Unlock()
	if create.Path != "/projects" || create.Body["name"] != "market" || create.Body["brief"] != "No scams" || create.Body["usedByMinors"] != false {
		t.Fatalf("create request = %+v", create)
	}

	if _, err := exec.LookPath("git"); err == nil {
		gitignore, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if !strings.Contains(string(gitignore), ".env.local") {
			t.Fatalf(".gitignore does not list .env.local: %q", gitignore)
		}
	}
}

func TestInitKeepsExistingProject(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("MODAPI_SECRET_KEY=sk_proj_existing\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := run(t, "init", "--dir", dir, "--server-url", api.URL, "-o", "json")
	if code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"status":"already_set_up"`) && !strings.Contains(stdout, `"status": "already_set_up"`) {
		t.Fatalf("expected already_set_up: %s", stdout)
	}
	for _, p := range api.paths() {
		if p == "POST /projects" {
			t.Fatal("init created a project although a key exists")
		}
	}
}

func TestInitWithoutCredentialsExplainsLogin(t *testing.T) {
	dir := setup(t)
	_, stderr, code := run(t, "init", "--dir", dir, "--agent-mode")
	if code != 3 {
		t.Fatalf("exit %d, want 3\nstderr: %s", code, stderr)
	}
	var envelope struct {
		ErrorType string   `json:"error_type"`
		Hints     []string `json:"hints"`
	}
	if err := json.Unmarshal([]byte(stderr), &envelope); err != nil {
		t.Fatalf("stderr is not an error envelope: %v\n%s", err, stderr)
	}
	if envelope.ErrorType != "authentication_error" || len(envelope.Hints) == 0 || !strings.Contains(envelope.Hints[0], "modapi login") {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestInitDryRunTouchesNothing(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")
	stdout, _, code := run(t, "init", "--dir", dir, "--server-url", api.URL, "--dry-run", "-o", "json")
	if code != 0 || !strings.Contains(stdout, `"would_create_project":true`) {
		t.Fatalf("exit %d: %s", code, stdout)
	}
	if len(api.paths()) != 0 {
		t.Fatalf("dry run called the API: %v", api.paths())
	}
	if _, err := os.Stat(filepath.Join(dir, ".env.local")); err == nil {
		t.Fatal("dry run wrote the env file")
	}
}

func TestDoctorReportsChecks(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("MODAPI_SECRET_KEY=sk_proj_new\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := run(t, "doctor", "--dir", dir, "--server-url", api.URL, "-o", "json")
	var result struct {
		OK     bool `json:"ok"`
		Checks []struct {
			ID, Status string
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s\n%s", err, stdout, stderr)
	}
	status := map[string]string{}
	for _, c := range result.Checks {
		status[c.ID] = c.Status
	}
	if status["project_key"] != "ok" || status["api"] != "ok" || status["sdk"] != "warn" || status["webhook_secret"] != "warn" {
		t.Fatalf("checks = %v", status)
	}
	if _, err := exec.LookPath("git"); err == nil {
		// .env.local is not in .gitignore in this fresh repo.
		if status["gitignore"] != "fail" || result.OK || code != 1 {
			t.Fatalf("gitignore = %s, ok = %v, exit = %d", status["gitignore"], result.OK, code)
		}
	}
}

func TestDoctorFailsWithoutKey(t *testing.T) {
	dir := setup(t)
	stdout, _, code := run(t, "doctor", "--dir", dir, "-o", "json")
	if code != 1 || !strings.Contains(stdout, `"project_key"`) {
		t.Fatalf("exit %d: %s", code, stdout)
	}
}
