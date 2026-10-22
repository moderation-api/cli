package custom

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// authServer holds the endpoints from the issuer's OAuth metadata.
type authServer struct {
	AuthorizationEndpoint       string `json:"authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
}

// tokenSet is the token endpoint's success response.
type tokenSet struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// oauthError is an OAuth error response (RFC 6749 section 5.2).
type oauthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

func (e *oauthError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Description)
	}
	return e.Code
}

// discover reads the authorization server metadata (RFC 8414).
func discover(ctx context.Context, client *http.Client, issuer string) (*authServer, error) {
	metadataURL := strings.TrimSuffix(issuer, "/") + "/.well-known/oauth-authorization-server"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach the sign-in server: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the sign-in server returned %s for its metadata", res.Status)
	}
	var server authServer
	if err := json.NewDecoder(res.Body).Decode(&server); err != nil {
		return nil, fmt.Errorf("could not read the sign-in server metadata: %w", err)
	}
	if server.AuthorizationEndpoint == "" || server.TokenEndpoint == "" {
		return nil, errors.New("the sign-in server metadata has no authorization or token endpoint")
	}
	return &server, nil
}

// pkce is a PKCE verifier and its S256 challenge (RFC 7636).
type pkce struct {
	Verifier  string
	Challenge string
}

func newPKCE() (pkce, error) {
	verifier, err := randomString(32)
	if err != nil {
		return pkce{}, err
	}
	sum := sha256.Sum256([]byte(verifier))
	return pkce{Verifier: verifier, Challenge: base64.RawURLEncoding.EncodeToString(sum[:])}, nil
}

func randomString(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// browserLoginOptions configures the authorization code flow with PKCE.
type browserLoginOptions struct {
	Server   *authServer
	ClientID string
	Resource string
	Scopes   []string
	// Open opens the authorization URL. When it fails, the URL is printed for the user to open.
	Open func(url string) error
	// Prompt receives the line that tells the user what to do.
	Prompt io.Writer
	Wait   time.Duration
	HTTP   *http.Client
}

// browserLogin runs the authorization code flow with PKCE and a one-shot
// loopback redirect (RFC 8252): the redirect URI is registered as
// http://127.0.0.1/callback and the port is picked at run time.
func browserLogin(ctx context.Context, opts browserLoginOptions) (*tokenSet, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("could not start the local sign-in callback: %w", err)
	}
	defer listener.Close()
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", listener.Addr().(*net.TCPAddr).Port)

	challenge, err := newPKCE()
	if err != nil {
		return nil, err
	}
	state, err := randomString(16)
	if err != nil {
		return nil, err
	}

	query := url.Values{
		"response_type":         {"code"},
		"client_id":             {opts.ClientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {strings.Join(opts.Scopes, " ")},
		"code_challenge":        {challenge.Challenge},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}
	if opts.Resource != "" {
		query.Set("resource", opts.Resource)
	}
	authURL := opts.Server.AuthorizationEndpoint + "?" + query.Encode()

	codes := make(chan callbackResult, 1)
	server := &http.Server{
		Handler:           callbackHandler(state, codes),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() { _ = server.Serve(listener) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if opts.Open == nil || opts.Open(authURL) != nil {
		fmt.Fprintf(opts.Prompt, "Open this URL in your browser to sign in:\n\n  %s\n\n", authURL)
	} else {
		fmt.Fprintf(opts.Prompt, "Opened your browser to sign in. If it did not open, go to:\n\n  %s\n\n", authURL)
	}

	waitCtx, cancel := context.WithTimeout(ctx, opts.Wait)
	defer cancel()
	var result callbackResult
	select {
	case result = <-codes:
	case <-waitCtx.Done():
		return nil, fmt.Errorf("sign-in did not finish within %s", opts.Wait)
	}
	if result.err != nil {
		return nil, result.err
	}

	return postToken(ctx, opts.HTTP, opts.Server.TokenEndpoint, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {result.code},
		"redirect_uri":  {redirectURI},
		"client_id":     {opts.ClientID},
		"code_verifier": {challenge.Verifier},
	})
}

type callbackResult struct {
	code string
	err  error
}

// callbackHandler answers the browser first and then hands the result to the
// CLI, so the browser never hangs on the redirect.
func callbackHandler(state string, results chan<- callbackResult) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		var result callbackResult
		switch {
		case query.Get("state") != state:
			result.err = errors.New("the sign-in response did not match this request; run the command again")
		case query.Get("error") != "":
			result.err = &oauthError{Code: query.Get("error"), Description: query.Get("error_description")}
		case query.Get("code") == "":
			result.err = errors.New("the sign-in response had no authorization code")
		default:
			result.code = query.Get("code")
		}
		writeCallbackPage(w, result.err)
		select {
		case results <- result:
		default:
		}
	})
	return mux
}

func writeCallbackPage(w http.ResponseWriter, err error) {
	title, body := "Signed in", "You are signed in to the Moderation API CLI. You can close this tab."
	status := http.StatusOK
	if err != nil {
		title, body = "Sign-in failed", "Sign-in did not finish: "+err.Error()+". Return to your terminal."
		status = http.StatusBadRequest
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	fmt.Fprintf(w, `<!doctype html><html><head><meta charset="utf-8"><title>%[1]s</title>
<style>body{font-family:system-ui,sans-serif;display:grid;place-items:center;min-height:100vh;margin:0;background:#0b0d10;color:#e6e8eb}
main{max-width:28rem;padding:2rem;text-align:center}h1{font-size:1.25rem;font-weight:600}p{color:#9aa3ad;line-height:1.5}</style>
</head><body><main><h1>%[1]s</h1><p>%[2]s</p></main></body></html>`, html.EscapeString(title), html.EscapeString(body))
}

// deviceLoginOptions configures the device authorization grant (RFC 8628).
type deviceLoginOptions struct {
	Server   *authServer
	ClientID string
	Resource string
	Scopes   []string
	Prompt   io.Writer
	HTTP     *http.Client
	// Sleep waits between polls; tests replace it.
	Sleep func(context.Context, time.Duration) error
}

type deviceAuthorization struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// deviceLogin signs in on a second device, for machines without a browser.
func deviceLogin(ctx context.Context, opts deviceLoginOptions) (*tokenSet, error) {
	if opts.Server.DeviceAuthorizationEndpoint == "" {
		return nil, errors.New("the sign-in server does not support device sign-in; run `modapi login` on a machine with a browser")
	}
	form := url.Values{
		"client_id": {opts.ClientID},
		"scope":     {strings.Join(opts.Scopes, " ")},
	}
	if opts.Resource != "" {
		form.Set("resource", opts.Resource)
	}
	var device deviceAuthorization
	if err := postForm(ctx, opts.HTTP, opts.Server.DeviceAuthorizationEndpoint, form, &device); err != nil {
		return nil, err
	}
	if device.DeviceCode == "" || device.UserCode == "" {
		return nil, errors.New("the sign-in server returned no device code")
	}

	verification := device.VerificationURIComplete
	if verification == "" {
		verification = device.VerificationURI
	}
	fmt.Fprintf(opts.Prompt, "To sign in, open this URL on any device:\n\n  %s\n\nand confirm the code: %s\n\n", verification, device.UserCode)

	interval := time.Duration(max(device.Interval, 5)) * time.Second
	expires := time.Duration(max(device.ExpiresIn, 60)) * time.Second
	deadline := time.Now().Add(expires)
	sleep := opts.Sleep
	if sleep == nil {
		sleep = sleepContext
	}

	for {
		if time.Now().After(deadline) {
			return nil, errors.New("the sign-in code expired; run `modapi login --device` again")
		}
		if err := sleep(ctx, interval); err != nil {
			return nil, err
		}
		tokens, err := postToken(ctx, opts.HTTP, opts.Server.TokenEndpoint, url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code": {device.DeviceCode},
			"client_id":   {opts.ClientID},
		})
		if err == nil {
			return tokens, nil
		}
		var oauthErr *oauthError
		if !errors.As(err, &oauthErr) {
			return nil, err
		}
		switch oauthErr.Code {
		case "authorization_pending":
		case "slow_down":
			interval += 5 * time.Second
		case "access_denied":
			return nil, errors.New("sign-in was declined")
		case "expired_token":
			return nil, errors.New("the sign-in code expired; run `modapi login --device` again")
		default:
			return nil, err
		}
	}
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func postToken(ctx context.Context, client *http.Client, endpoint string, form url.Values) (*tokenSet, error) {
	var tokens tokenSet
	if err := postForm(ctx, client, endpoint, form, &tokens); err != nil {
		return nil, err
	}
	if tokens.AccessToken == "" {
		return nil, errors.New("the sign-in server returned no access token")
	}
	return &tokens, nil
}

// postForm posts a form and decodes a JSON success body into out. An OAuth
// error body comes back as *oauthError.
func postForm(ctx context.Context, client *http.Client, endpoint string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach the sign-in server: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		var oauthErr oauthError
		if json.Unmarshal(body, &oauthErr) == nil && oauthErr.Code != "" {
			return &oauthErr
		}
		return fmt.Errorf("the sign-in server returned %s", res.Status)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("could not read the sign-in server response: %w", err)
	}
	return nil
}
