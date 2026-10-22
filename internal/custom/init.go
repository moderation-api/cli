package custom

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moderation-api/cli/internal/forms"
	"github.com/moderation-api/cli/internal/output"
	"github.com/moderation-api/cli/internal/sdk/models/operations"
	"github.com/moderation-api/cli/internal/sdk/models/sdkerrors"
	"github.com/moderation-api/cli/internal/usage"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up Moderation API in this project",
		Long: `Set up Moderation API in the project in the current directory.

init detects the language, framework and package manager, creates a
Moderation API project, and writes its secret key (MODAPI_SECRET_KEY) and
webhook signing secret (MODAPI_WEBHOOK_SECRET) to the project's env file. It
adds the env file to .gitignore when git would track it.

Describe the platform with --domain, --context and --brief: the project uses
them to suggest moderation policies. Coding agents can read these from the
codebase and confirm them with the user before running init.

init needs an organization key to create the project: sign in with
"modapi login", or set MODAPI_ORGANIZATION_KEY. When the env file already has
MODAPI_SECRET_KEY, init checks it and creates nothing (use --force for a new
project).

init does not change your code. Use the moderation-api-setup agent skill, or
the guide at https://docs.moderationapi.com, to add the moderation calls.`,
		Example: `  modapi init
  modapi init --name "Acme Marketplace" --domain acme.com \
    --context "Marketplace where buyers and sellers chat before a booking" \
    --brief "Stop contact details and payment requests outside the platform, scams and harassment"
  modapi init --dry-run -o json`,
		Args: cobra.NoArgs,
		RunE: runInit,
	}
	f := cmd.Flags()
	f.StringP("name", "n", "", "Project name (default: the package or directory name)")
	f.String("domain", "", "The website or app the project moderates content for, e.g. acme.com")
	f.StringP("context", "c", "", "What the platform is and who uses it")
	f.StringP("brief", "b", "", "What to moderate, the problems you see and your guidelines")
	f.Bool("used-by-minors", false, "Children use the platform (stricter policies)")
	f.String("dir", ".", "The project directory")
	f.String("env-file", "", "Env file to write the secrets to (default: .env.local for Next.js, else .env)")
	f.Bool("install", false, "Also install the SDK with the detected package manager")
	f.Bool("force", false, "Create a new project even when the env file already has a key")
	f.String("organization-key", "", "Organization key to create the project with. Prefer the MODAPI_ORGANIZATION_KEY env var: a flag value stays in shell history (default: the key from modapi login)")
	return cmd
}

type initProject struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Domain  string `json:"domain,omitempty"`
	Context string `json:"context,omitempty"`
}

type initResult struct {
	Status          string       `json:"status"` // "created" or "already_set_up"
	Project         *initProject `json:"project,omitempty"`
	EnvFile         string       `json:"env_file"`
	VariablesSet    []string     `json:"variables_set"`
	GitignoreAdded  bool         `json:"gitignore_added"`
	Stack           *stack       `json:"stack,omitempty"`
	SDKInstalledNow bool         `json:"sdk_installed_now"`
	DashboardURL    string       `json:"dashboard_url,omitempty"`
	Warnings        []string     `json:"warnings,omitempty"`
	NextSteps       []string     `json:"next_steps"`
}

type initInput struct {
	Name         string `json:"name"`
	Domain       string `json:"domain,omitempty"`
	Context      string `json:"context,omitempty"`
	Brief        string `json:"brief,omitempty"`
	UsedByMinors bool   `json:"used_by_minors"`
}

func runInit(cmd *cobra.Command, _ []string) error {
	if usage.UsageRequested(cmd) {
		return usage.EmitSchema(cmd, cmd.OutOrStdout())
	}
	dir, _ := cmd.Flags().GetString("dir")
	dir, err := filepath.Abs(dir)
	if err != nil {
		return failure(output.ReasonCLIValidation, "invalid --dir: %v", err)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return failure(output.ReasonCLIValidation, "--dir %s is not a directory", dir)
	}

	detected := detectStack(dir)
	envFile := chooseEnvFile(cmd, dir, detected)
	input := initInputFromFlags(cmd, dir)
	force, _ := cmd.Flags().GetBool("force")
	install, _ := cmd.Flags().GetBool("install")

	existingKey := ""
	if !force {
		existingKey, envFile = existingProjectKey(dir, detected, envFile)
	}

	if dryRunPlan(cmd, "init makes no request and writes no files under --dry-run; this is what it would do.", map[string]any{
		"directory":               dir,
		"stack":                   detected,
		"env_file":                envFile,
		"existing_key_found":      existingKey != "",
		"would_create_project":    existingKey == "",
		"project":                 input,
		"would_install_sdk":       install && detected != nil && !detected.SDKInstalled && detected.InstallCommand != "",
		"organization_key_source": organizationKeySource(cmd),
	}) {
		return nil
	}

	result := initResult{EnvFile: relPath(dir, envFile), Stack: detected, VariablesSet: []string{}}

	if existingKey != "" {
		if err := checkExistingKey(cmd, existingKey, &result); err != nil {
			return err
		}
	} else {
		if err := createProject(cmd, dir, envFile, input, &result); err != nil {
			return err
		}
	}

	if install && detected != nil && !detected.SDKInstalled && detected.InstallCommand != "" {
		if err := runInstall(cmd, dir, detected.InstallCommand); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Installing the SDK failed: %v. Run `%s` yourself.", err, detected.InstallCommand))
		} else {
			result.SDKInstalledNow = true
		}
	}
	result.NextSteps = nextSteps(detected, &result)
	return render(cmd, result, func(w io.Writer) { printInitResult(w, &result) })
}

func chooseEnvFile(cmd *cobra.Command, dir string, s *stack) string {
	if name, _ := cmd.Flags().GetString("env-file"); name != "" {
		if filepath.IsAbs(name) {
			return name
		}
		return filepath.Join(dir, name)
	}
	if s != nil && s.EnvFile != "" {
		return filepath.Join(dir, s.EnvFile)
	}
	return filepath.Join(dir, ".env")
}

// existingProjectKey finds a key already in the project's env files. When
// found, it returns the file it is in, so init reports and keeps that file.
func existingProjectKey(dir string, s *stack, envFile string) (string, string) {
	if v := readEnvValue(envFile, projectKeyEnv); v != "" {
		return v, envFile
	}
	if v, path := findEnvValue(dir, s, projectKeyEnv); v != "" {
		return v, path
	}
	return "", envFile
}

func initInputFromFlags(cmd *cobra.Command, dir string) initInput {
	f := cmd.Flags()
	var in initInput
	in.Name, _ = f.GetString("name")
	in.Domain, _ = f.GetString("domain")
	in.Context, _ = f.GetString("context")
	in.Brief, _ = f.GetString("brief")
	in.UsedByMinors, _ = f.GetBool("used-by-minors")
	if in.Name == "" {
		in.Name = defaultProjectName(dir)
	}
	return in
}

// defaultProjectName is the package name, else the directory name, cut to the API's 35 characters.
func defaultProjectName(dir string) string {
	name := filepath.Base(dir)
	var manifest struct {
		Name string `json:"name"`
	}
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil && json.Unmarshal(data, &manifest) == nil && manifest.Name != "" {
		// "@acme/web" becomes "web".
		name = manifest.Name[strings.LastIndex(manifest.Name, "/")+1:]
	}
	if runes := []rune(name); len(runes) > 35 {
		name = string(runes[:35])
	}
	return name
}

// organizationKeySource names where init would take the organization key
// from, without reading the keychain (it runs under --dry-run).
func organizationKeySource(cmd *cobra.Command) string {
	if key, source := explicitOrganizationKey(cmd); key != "" {
		return source
	}
	return "login session, if signed in"
}

func explicitOrganizationKey(cmd *cobra.Command) (key, source string) {
	if v, _ := cmd.Flags().GetString("organization-key"); v != "" {
		return v, "flag"
	}
	if v := strings.TrimSpace(os.Getenv(organizationKeyEnv)); v != "" {
		return v, "env"
	}
	return "", ""
}

// organizationKey resolves the key that creates projects: flag, env var, then the login session.
func organizationKey(cmd *cobra.Command) (key, source string, err error) {
	if key, source := explicitOrganizationKey(cmd); key != "" {
		return key, source, nil
	}
	s, err := loadSession()
	if err != nil {
		return "", "unset", err
	}
	if s != nil {
		return s.APIKey, "login", nil
	}
	return "", "unset", nil
}

func checkExistingKey(cmd *cobra.Command, key string, result *initResult) error {
	result.Status = "already_set_up"
	api, err := newAPIClient(cmd, key)
	if err != nil {
		return err
	}
	res, err := api.Account.Retrieve(cmd.Context(), &operations.GetAccountRequest{})
	if err != nil {
		var authErr *sdkerrors.AuthenticationError
		var permErr *sdkerrors.PermissionDeniedError
		if errors.As(err, &authErr) || errors.As(err, &permErr) {
			return withHints(failure(output.ReasonCLIAuthentication, "%s in %s was rejected by the API: %v", projectKeyEnv, result.EnvFile, err),
				"Replace it with a valid project key from the dashboard's API keys page",
				"Or run `modapi init --force` to create a new project")
		}
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check %s in %s: %v. Run `modapi doctor` later.", projectKeyEnv, result.EnvFile, err))
		return nil
	}
	if res.Object != nil && res.Object.CurrentProject != nil {
		p := res.Object.CurrentProject
		result.Project = &initProject{ID: p.ID, Name: p.Name}
		result.DashboardURL = fmt.Sprintf("%s/project/%s/policies", loadSettings().DashboardURL, p.ID)
	}
	return nil
}

func createProject(cmd *cobra.Command, dir, envFile string, input initInput, result *initResult) error {
	// Refuse before creating anything if the secrets would land in a committed file.
	switch gitIgnoreState(dir, envFile) {
	case "tracked":
		name := relPath(dir, envFile)
		return withHints(failure(output.ReasonCLIValidation, "%s is committed to git, so writing the project key there would commit it", name),
			"Stop tracking it: git rm --cached "+name+", then add it to .gitignore",
			"Or write the keys to another file with --env-file .env.local")
	case "unknown":
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not tell whether git ignores %s. Make sure it is not committed.", relPath(dir, envFile)))
	}

	orgKey, _, err := organizationKey(cmd)
	if err != nil {
		return failure(output.ReasonCLIAuthentication, "%v", err)
	}
	if orgKey == "" {
		return withHints(failure(output.ReasonCLIAuthentication, "init needs an organization key to create a project"),
			"Run `modapi login` to sign in in the browser (`modapi login --device` on a machine without a browser)",
			"Or set "+organizationKeyEnv+" to an organization key from the dashboard's API keys page",
			"To use a project you already have, put its key in the env file as "+projectKeyEnv+" and run `modapi init` again")
	}

	if canPrompt(cmd) && !cmd.Flags().Changed("context") && !cmd.Flags().Changed("brief") {
		if err := promptProject(cmd, &input); err != nil {
			return err
		}
	}
	if input.Domain == "" && input.Context == "" {
		result.Warnings = append(result.Warnings, "No --domain or --context was given, so the project suggests no policies. Add context in the dashboard, or run `modapi projects update` with --domain or --context.")
	}

	// Not the --debug client: it would log the new project's secret key.
	api, err := newQuietAPIClient(cmd, orgKey)
	if err != nil {
		return err
	}
	req := operations.CreateProjectRequest{Name: input.Name, UsedByMinors: &input.UsedByMinors}
	if input.Domain != "" {
		req.Domain = &input.Domain
	}
	if input.Context != "" {
		req.Context = &input.Context
	}
	if input.Brief != "" {
		req.Brief = &input.Brief
	}
	stop := output.StartRequestProgress(cmd)
	res, err := api.Projects.Create(cmd.Context(), req)
	stop()
	if err != nil {
		return output.Error(cmd, err)
	}
	if res.ProjectWithKey == nil || res.ProjectWithKey.SecretKey == "" {
		return failure(output.ReasonCLIProtocol, "the API created the project but returned no secret key")
	}
	project := res.ProjectWithKey

	values := map[string]string{projectKeyEnv: project.SecretKey}
	order := []string{projectKeyEnv, webhookSecretEnv}
	if secret, err := webhookSecret(cmd, project.SecretKey); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not get the webhook signing secret: %v. Run `modapi webhook-secret retrieve` later.", err))
	} else {
		values[webhookSecretEnv] = secret
	}
	added, err := ensureGitIgnored(dir, envFile)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not add %s to .gitignore: %v. Make sure it is not committed.", result.EnvFile, err))
	}
	result.GitignoreAdded = added
	if err := setEnvValues(envFile, values, order); err != nil {
		return failure(output.ReasonCLIRuntime, "the project %s was created, but writing %s failed: %v. Find the key in the dashboard", project.ID, envFile, err)
	}
	for _, key := range order {
		if _, ok := values[key]; ok {
			result.VariablesSet = append(result.VariablesSet, key)
		}
	}

	result.Status = "created"
	result.Project = &initProject{ID: project.ID, Name: project.Name, Domain: deref(project.Domain), Context: deref(project.Context)}
	result.DashboardURL = fmt.Sprintf("%s/project/%s/policies", loadSettings().DashboardURL, project.ID)
	return nil
}

func webhookSecret(cmd *cobra.Command, projectKey string) (string, error) {
	api, err := newQuietAPIClient(cmd, projectKey)
	if err != nil {
		return "", err
	}
	res, err := api.WebhookSecret.Retrieve(cmd.Context(), &operations.GetWebhookSecretRequest{})
	if err != nil {
		return "", err
	}
	if res.Object == nil || res.Object.Secret == "" {
		return "", fmt.Errorf("the API returned no secret")
	}
	return res.Object.Secret, nil
}

// promptProject asks a person at a terminal for what agents pass as flags.
func promptProject(cmd *cobra.Command, in *initInput) error {
	fields := []*forms.Field{
		forms.NewInput(&in.Name).Title("Project name").Placeholder(in.Name),
		forms.NewInput(&in.Domain).Title("Website or app").Description("e.g. acme.com. Used to suggest policies.").Placeholder(in.Domain),
		forms.NewText(&in.Context).Title("What is the platform, and who uses it?").Description("Optional. e.g. a marketplace where buyers and sellers chat"),
		forms.NewText(&in.Brief).Title("What should be moderated?").Description("Optional. The problems you see and your guidelines"),
		forms.NewConfirm(&in.UsedByMinors).Title("Do children use the platform?"),
	}
	name := in.Name
	form := forms.New(forms.NewPage("Set up Moderation API", fields...)).IO(cmd.InOrStdin(), cmd.ErrOrStderr())
	if err := form.Run(); err != nil {
		return fmt.Errorf("init: %w", err)
	}
	if strings.TrimSpace(in.Name) == "" {
		in.Name = name
	}
	return nil
}

func runInstall(cmd *cobra.Command, dir, command string) error {
	parts := strings.Fields(command)
	install := exec.CommandContext(cmd.Context(), parts[0], parts[1:]...)
	install.Dir = dir
	// The package manager's output is progress, not the command's result.
	install.Stdout = cmd.ErrOrStderr()
	install.Stderr = cmd.ErrOrStderr()
	return install.Run()
}

func nextSteps(s *stack, r *initResult) []string {
	var steps []string
	if s != nil && !s.SDKInstalled && !r.SDKInstalledNow && s.InstallCommand != "" {
		steps = append(steps, "Install the SDK: "+s.InstallCommand)
	}
	if s != nil && s.Note != "" {
		steps = append(steps, s.Note)
	}
	steps = append(steps,
		"Moderate content where users create it: call content.submit and act on recommendation.action (allow, review or reject). The moderation-api-setup agent skill does this for you: npx skills add moderation-api/cli",
		"Restart your dev server so it reads the new env values.",
	)
	if r.DashboardURL != "" {
		steps = append(steps, "Review the suggested policies: "+r.DashboardURL)
	}
	steps = append(steps, "Check the setup any time: modapi doctor")
	return steps
}

func printInitResult(w io.Writer, r *initResult) {
	if r.Stack != nil {
		desc := r.Stack.Language
		if r.Stack.Framework != "" {
			desc = r.Stack.Framework + " (" + desc + ")"
		}
		fmt.Fprintf(w, "Detected %s\n", desc)
	}
	switch r.Status {
	case "created":
		fmt.Fprintf(w, "Created project %s (%s)\n", r.Project.Name, r.Project.ID)
		fmt.Fprintf(w, "Wrote %s to %s\n", strings.Join(r.VariablesSet, " and "), r.EnvFile)
		if r.GitignoreAdded {
			fmt.Fprintf(w, "Added %s to .gitignore\n", r.EnvFile)
		}
	case "already_set_up":
		if r.Project != nil {
			fmt.Fprintf(w, "Already set up: %s in %s is for project %s (%s)\n", projectKeyEnv, r.EnvFile, r.Project.Name, r.Project.ID)
		} else {
			fmt.Fprintf(w, "%s is already in %s\n", projectKeyEnv, r.EnvFile)
		}
	}
	if r.SDKInstalledNow {
		fmt.Fprintf(w, "Installed %s\n", r.Stack.SDKPackage)
	}
	for _, warning := range r.Warnings {
		fmt.Fprintf(w, "\nWarning: %s\n", warning)
	}
	fmt.Fprintln(w, "\nNext steps:")
	for i, step := range r.NextSteps {
		fmt.Fprintf(w, "  %d. %s\n", i+1, step)
	}
}

func relPath(dir, path string) string {
	if rel, err := filepath.Rel(dir, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
