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
	// projects is the JSON array GET /projects returns (default: none).
	projects string
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
		case r.Method == http.MethodGet && r.URL.Path == "/projects":
			projects := f.projects
			if projects == "" {
				projects = "[]"
			}
			_, _ = io.WriteString(w, `{"data":`+projects+`,"pagination":{"page":1,"pageSize":100,"total":0,"totalPages":1,"hasMore":false}}`)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/api-keys"):
			if auth != "Bearer sk_org_test" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"message":"bad key","code":"unauthorized"}`)
				return
			}
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/projects/"), "/api-keys")
			_, _ = io.WriteString(w, `{"id":"key_1","name":"`+body["name"].(string)+`","projectId":"`+id+`","createdAt":"2026-10-23T00:00:00Z","secretKey":"sk_proj_linked"}`)
		case r.URL.Path == "/webhook-secret":
			_, _ = io.WriteString(w, `{"secret":"whsec_new"}`)
		case r.URL.Path == "/account":
			if auth != "Bearer sk_proj_new" && auth != "Bearer sk_proj_existing" && auth != "Bearer sk_proj_linked" {
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

	if paths := api.paths(); len(paths) < 2 || paths[0] != "GET /projects" || paths[1] != "POST /projects" {
		t.Fatalf("requests = %v, want GET /projects (none to link), then POST /projects", paths)
	}
	api.mu.Lock()
	create := api.requests[1]
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
	stdout, _, code := run(t, "init", "--dir", dir, "--server-url", api.URL, "--new", "--dry-run", "-o", "json")
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

func TestInitRefusesTrackedEnvFile(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := setup(t)
	api := newFakeAPI(t)
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")
	env := filepath.Join(dir, ".env.local")
	if err := os.WriteFile(env, []byte("OTHER=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", ".env.local"}, {"-c", "user.email=t@t", "-c", "user.name=t", "commit", "-qm", "x"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	_, stderr, code := run(t, "init", "--dir", dir, "--server-url", api.URL, "--agent-mode")
	if code != 2 || !strings.Contains(stderr, "committed to git") {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if len(api.paths()) != 0 {
		t.Fatalf("init created a project before refusing: %v", api.paths())
	}
	if data, _ := os.ReadFile(env); strings.Contains(string(data), "MODAPI_SECRET_KEY") {
		t.Fatal("init wrote the key into a tracked file")
	}
}

func TestInitRejectsInvalidExistingKey(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("MODAPI_SECRET_KEY=sk_revoked\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, stderr, code := run(t, "init", "--dir", dir, "--server-url", api.URL, "--agent-mode")
	if code != 3 || !strings.Contains(stderr, "--force") {
		t.Fatalf("exit %d: %s", code, stderr)
	}
}

func (f *fakeAPI) authFor(path string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, r := range f.requests {
		if r.Path == path {
			out = append(out, r.Auth)
		}
	}
	return out
}

func writeEnv(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestAPICommandUsesProjectKeyFromEnvFile(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	writeEnv(t, dir, ".env", "MODAPI_SECRET_KEY=sk_proj_existing\n")
	t.Chdir(dir)

	stdout, stderr, code := run(t, "account", "retrieve", "--server-url", api.URL, "-o", "json")
	if code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	if got := api.authFor("/account"); len(got) != 1 || got[0] != "Bearer sk_proj_existing" {
		t.Fatalf("Authorization = %v", got)
	}
}

func TestAPICommandPrefersExplicitKey(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	writeEnv(t, dir, ".env", "MODAPI_SECRET_KEY=sk_proj_existing\n")
	t.Chdir(dir)
	t.Setenv("MODAPI_SECRET_KEY", "sk_proj_new")

	if _, stderr, code := run(t, "account", "retrieve", "--server-url", api.URL); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if got := api.authFor("/account"); len(got) != 1 || got[0] != "Bearer sk_proj_new" {
		t.Fatalf("Authorization = %v", got)
	}
}

func TestManagementCommandUsesOrganizationKey(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	writeEnv(t, dir, ".env", "MODAPI_SECRET_KEY=sk_proj_existing\n")
	t.Chdir(dir)
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")

	_, _, _ = run(t, "projects", "list", "--server-url", api.URL)
	if got := api.authFor("/projects"); len(got) != 1 || got[0] != "Bearer sk_org_test" {
		t.Fatalf("Authorization = %v", got)
	}
}

func TestAPICommandWithoutAnyKeyStopsBeforeTheRequest(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	t.Chdir(dir)

	_, stderr, code := run(t, "account", "retrieve", "--server-url", api.URL)
	if code != 3 {
		t.Fatalf("exit %d, want 3: %s", code, stderr)
	}
	if !strings.Contains(stderr, "modapi init") || !strings.Contains(stderr, "modapi login") {
		t.Fatalf("stderr does not say how to get a key: %s", stderr)
	}
	if paths := api.paths(); len(paths) != 0 {
		t.Fatalf("sent requests without a key: %v", paths)
	}
}

func TestDoctorListsProblemsWithoutAnErrorBlock(t *testing.T) {
	dir := setup(t)

	stdout, stderr, code := run(t, "doctor", "--dir", dir, "--agent-mode=false", "--color", "never")
	if code != 1 {
		t.Fatalf("exit %d, want 1\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	for _, want := range []string{"Problems\n  ✗ No project key found.\n    Fix: modapi init\n", "Warnings\n", "1 problem, 2 warnings.\n"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout misses %q:\n%s", want, stdout)
		}
	}
	if stderr != "" {
		t.Fatalf("doctor printed an error block: %s", stderr)
	}
}

const twoProjects = `[
  {"id":"proj_1","name":"Market","description":null,"domain":"acme.com","context":null,"globalFlaggingMode":"ACTIVE","enableRequestLogging":true,"casebookEnabled":true,"createdAt":"2026-10-22T00:00:00Z","updatedAt":"2026-10-22T00:00:00Z"},
  {"id":"proj_2","name":"Forum","description":null,"domain":null,"context":null,"globalFlaggingMode":"ACTIVE","enableRequestLogging":true,"casebookEnabled":true,"createdAt":"2026-10-22T00:00:00Z","updatedAt":"2026-10-22T00:00:00Z"}
]`

func readEnv(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestInitWithProjectsAsksAnAgentToChoose(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	api.projects = twoProjects
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")

	_, stderr, code := run(t, "init", "--dir", dir, "--server-url", api.URL, "-o", "json")
	if code != 2 {
		t.Fatalf("exit %d, want 2: %s", code, stderr)
	}
	for _, want := range []string{"Market (proj_1)", "Forum (proj_2)", "--project", "--new"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr misses %q: %s", want, stderr)
		}
	}
	if paths := api.paths(); len(paths) != 1 || paths[0] != "GET /projects" {
		t.Fatalf("requests = %v, want only the project list", paths)
	}
}

func TestInitLinksAProjectByName(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	api.projects = twoProjects
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")

	stdout, stderr, code := run(t, "init", "--dir", dir, "--server-url", api.URL, "--project", "forum", "-o", "json")
	if code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"status":"linked"`) || strings.Contains(stdout, "sk_proj_linked") {
		t.Fatalf("result = %s", stdout)
	}
	if paths := api.paths(); !strings.Contains(strings.Join(paths, ","), "POST /projects/proj_2/api-keys") {
		t.Fatalf("requests = %v", paths)
	}
	api.mu.Lock()
	var name string
	for _, r := range api.requests {
		if strings.HasSuffix(r.Path, "/api-keys") {
			name, _ = r.Body["name"].(string)
		}
	}
	api.mu.Unlock()
	if !strings.HasPrefix(name, "CLI · ") {
		t.Fatalf("key name = %q", name)
	}
	if env := readEnv(t, filepath.Join(dir, ".env.local")); env != "MODAPI_SECRET_KEY=sk_proj_linked\nMODAPI_WEBHOOK_SECRET=whsec_new\n" {
		t.Fatalf(".env.local = %q", env)
	}
}

func TestInitRefusesProjectAndNewTogether(t *testing.T) {
	dir := setup(t)
	_, _, code := run(t, "init", "--dir", dir, "--project", "x", "--new")
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
}

func TestLinkRefusesAmbiguousName(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	api.projects = strings.Replace(twoProjects, `"Forum"`, `"market"`, 1)
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")

	_, stderr, code := run(t, "link", "Market", "--dir", dir, "--server-url", api.URL)
	if code != 2 || !strings.Contains(stderr, "more than one project") {
		t.Fatalf("exit %d: %s", code, stderr)
	}
}

func TestLinkAsksBeforeReplacingAnotherProjectsKey(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	api.projects = twoProjects
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")
	// The account endpoint says sk_proj_existing belongs to proj_1.
	writeEnv(t, dir, ".env.local", "MODAPI_SECRET_KEY=sk_proj_existing\n")

	_, stderr, code := run(t, "link", "proj_2", "--dir", dir, "--server-url", api.URL)
	if code != 2 || !strings.Contains(stderr, "--force") {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if env := readEnv(t, filepath.Join(dir, ".env.local")); !strings.Contains(env, "sk_proj_existing") {
		t.Fatalf("the key was replaced without --force: %q", env)
	}

	stdout, stderr, code := run(t, "link", "proj_2", "--dir", dir, "--server-url", api.URL, "--force", "-o", "json")
	if code != 0 || !strings.Contains(stdout, `"status":"linked"`) {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	if env := readEnv(t, filepath.Join(dir, ".env.local")); !strings.Contains(env, "MODAPI_SECRET_KEY=sk_proj_linked") {
		t.Fatalf(".env.local = %q", env)
	}
}

func TestLinkSameProjectChangesNothing(t *testing.T) {
	dir := setup(t)
	api := newFakeAPI(t)
	api.projects = twoProjects
	t.Setenv("MODAPI_ORGANIZATION_KEY", "sk_org_test")
	writeEnv(t, dir, ".env.local", "MODAPI_SECRET_KEY=sk_proj_existing\n")

	stdout, stderr, code := run(t, "link", "Market", "--dir", dir, "--server-url", api.URL, "-o", "json")
	if code != 0 || !strings.Contains(stdout, `"status":"already_set_up"`) {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	for _, p := range api.paths() {
		if strings.HasSuffix(p, "/api-keys") {
			t.Fatalf("made a key for the project it is already linked to: %v", api.paths())
		}
	}
}
