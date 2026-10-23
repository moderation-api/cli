package custom

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/moderation-api/cli/internal/clierrors"
	"github.com/moderation-api/cli/internal/config"
	"github.com/moderation-api/cli/internal/flagutil"
	"github.com/moderation-api/cli/internal/sdk/models/operations"
	"github.com/moderation-api/cli/internal/usage"
)

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the Moderation API setup of this project",
		Long: `Check the Moderation API setup of the project in the current directory.

doctor finds the project key (flag, MODAPI_SECRET_KEY, the project's env
files, then the keychain), checks it against the API, and checks the webhook
signing secret, .gitignore, the SDK dependency and the modapi login session.

Each check has a status (ok, warn, fail or skip) and, when it is not ok, a
remedy to run. doctor exits 1 when a check fails. Use -o json for the result
as data.`,
		Example: `  modapi doctor
  modapi doctor -o json`,
		Args: cobra.NoArgs,
		RunE: runDoctor,
	}
	cmd.Flags().String("dir", ".", "The project directory")
	return cmd
}

type check struct {
	ID      string `json:"id"`
	Status  string `json:"status"` // ok, warn, fail or skip
	Message string `json:"message"`
	Remedy  string `json:"remedy,omitempty"`
}

type doctorResult struct {
	OK     bool    `json:"ok"`
	Stack  *stack  `json:"stack,omitempty"`
	Checks []check `json:"checks"`
}

func runDoctor(cmd *cobra.Command, _ []string) error {
	if usage.UsageRequested(cmd) {
		return usage.EmitSchema(cmd, cmd.OutOrStdout())
	}
	if dryRunPlan(cmd, "doctor calls the API to check the key; nothing was sent under --dry-run.", nil) {
		return nil
	}
	dir, _ := cmd.Flags().GetString("dir")
	dir, _ = filepath.Abs(dir)
	detected := detectStack(dir)
	result := doctorResult{OK: true, Stack: detected}
	add := func(c check) {
		if c.Status == "fail" {
			result.OK = false
		}
		result.Checks = append(result.Checks, c)
	}

	key, source, envFile := findProjectKey(cmd, dir, detected)
	if key == "" {
		add(check{ID: "project_key", Status: "fail", Message: "No project key found.", Remedy: "modapi init"})
	} else {
		add(check{ID: "project_key", Status: "ok", Message: fmt.Sprintf("Project key %s from %s.", maskKey(key), source)})
		for _, c := range checkKeyAgainstAPI(cmd, key) {
			add(c)
		}
	}

	add(checkWebhookSecret(dir, detected))
	if envFile != "" {
		add(checkGitignore(dir, envFile))
	}
	add(checkSDK(detected))
	add(checkLogin())

	if err := render(cmd, result, func(w io.Writer) { printDoctor(w, &result) }); err != nil {
		return err
	}
	if !result.OK {
		failed := 0
		for _, c := range result.Checks {
			if c.Status == "fail" {
				failed++
			}
		}
		return clierrors.WithExitCode(withHints(fmt.Errorf("%d check(s) failed", failed),
			"Run the remedy listed for each failed check, then run `modapi doctor` again"), 1)
	}
	return nil
}

// findProjectKey resolves the project key like the SDKs and the CLI do, with
// the project's env files between the environment and the keychain.
func findProjectKey(cmd *cobra.Command, dir string, s *stack) (key, source, envFile string) {
	if v, changed := flagutil.GetStringFlag(cmd, "secret-key"); changed && v != "" {
		_, file := findEnvValue(dir, s, projectKeyEnv)
		return v, "--secret-key", file
	}
	if v := os.Getenv(projectKeyEnv); v != "" {
		_, file := findEnvValue(dir, s, projectKeyEnv)
		return v, projectKeyEnv + " in the environment", file
	}
	if v, file := findEnvValue(dir, s, projectKeyEnv); v != "" {
		return v, relPath(dir, file), file
	}
	if v, src := config.ResolveSecurityCredential(cmd, "secret-key"); v != "" {
		return v, "the CLI " + src + " (not the project's env file)", ""
	}
	return "", "", ""
}

func checkKeyAgainstAPI(cmd *cobra.Command, key string) []check {
	api, err := newAPIClient(cmd, key)
	if err != nil {
		return []check{{ID: "api", Status: "fail", Message: err.Error()}}
	}
	res, err := api.Account.Retrieve(cmd.Context(), &operations.GetAccountRequest{})
	if err != nil {
		return []check{{ID: "api", Status: "fail", Message: "The API did not accept the key: " + err.Error(), Remedy: "Check the key on the dashboard's API keys page, or run `modapi init --force` for a new project."}}
	}
	if res.Object == nil {
		return []check{{ID: "api", Status: "warn", Message: "The API accepted the key but returned no account details."}}
	}
	account := res.Object
	message := "The API accepted the key"
	if account.CurrentProject != nil {
		message += fmt.Sprintf(" for project %s (%s)", account.CurrentProject.Name, account.CurrentProject.ID)
	}
	checks := []check{{ID: "api", Status: "ok", Message: message + "."}}

	quota := check{ID: "quota", Status: "ok", Message: fmt.Sprintf("%.0f of %.0f requests left on the %s plan.", account.RemainingQuota, account.TextAPIQuota, account.PaidPlanName)}
	switch {
	case account.TextAPIQuota > 0 && account.RemainingQuota <= 0:
		quota.Status, quota.Remedy = "fail", "Upgrade the plan on the dashboard's billing page."
		quota.Message = "No requests left this period."
	case account.TextAPIQuota > 0 && account.RemainingQuota < account.TextAPIQuota*0.1:
		quota.Status, quota.Remedy = "warn", "Upgrade the plan on the dashboard's billing page."
	}
	return append(checks, quota)
}

func checkWebhookSecret(dir string, s *stack) check {
	if os.Getenv(webhookSecretEnv) != "" {
		return check{ID: "webhook_secret", Status: "ok", Message: webhookSecretEnv + " is set in the environment."}
	}
	if _, file := findEnvValue(dir, s, webhookSecretEnv); file != "" {
		return check{ID: "webhook_secret", Status: "ok", Message: webhookSecretEnv + " is in " + relPath(dir, file) + "."}
	}
	return check{
		ID: "webhook_secret", Status: "warn",
		Message: webhookSecretEnv + " is not set. You need it only to verify webhooks.",
		Remedy:  "modapi webhook-secret retrieve, then add the secret as " + webhookSecretEnv,
	}
}

func checkGitignore(dir, envFile string) check {
	name := relPath(dir, envFile)
	switch gitIgnoreState(dir, envFile) {
	case "tracked":
		return check{ID: "gitignore", Status: "fail", Message: name + " holds secrets and is committed to git.", Remedy: "git rm --cached " + name + ", add it to .gitignore, and rotate the keys on the dashboard's API keys page."}
	case "unknown":
		return check{ID: "gitignore", Status: "warn", Message: "Could not tell whether git ignores " + name + "."}
	case "ignored":
		return check{ID: "gitignore", Status: "ok", Message: name + " is in .gitignore."}
	case "not_ignored":
		return check{ID: "gitignore", Status: "fail", Message: name + " holds secrets but git does not ignore it.", Remedy: "Add " + name + " to .gitignore."}
	default:
		return check{ID: "gitignore", Status: "skip", Message: "Not a git repository."}
	}
}

func checkSDK(s *stack) check {
	switch {
	case s == nil:
		return check{ID: "sdk", Status: "skip", Message: "No project manifest found (package.json, pyproject.toml, Gemfile, composer.json or go.mod)."}
	case s.SDKPackage == "":
		return check{ID: "sdk", Status: "skip", Message: s.Note}
	case s.SDKInstalled:
		return check{ID: "sdk", Status: "ok", Message: s.SDKPackage + " is a dependency."}
	default:
		return check{ID: "sdk", Status: "warn", Message: s.SDKPackage + " is not a dependency.", Remedy: s.InstallCommand}
	}
}

func checkLogin() check {
	s, err := loadSession()
	switch {
	case err != nil:
		return check{ID: "login", Status: "warn", Message: err.Error(), Remedy: "modapi login"}
	case s == nil:
		return check{ID: "login", Status: "skip", Message: "Not signed in. You need to sign in only to create projects with modapi init.", Remedy: "modapi login"}
	default:
		return check{ID: "login", Status: "ok", Message: "Signed in to " + s.OrganizationName + "."}
	}
}

func printDoctor(w io.Writer, r *doctorResult) {
	marks := map[string]string{"ok": "✓", "warn": "!", "fail": "✗", "skip": "-"}
	for _, c := range r.Checks {
		fmt.Fprintf(w, "%s %s\n", marks[c.Status], c.Message)
		if c.Remedy != "" {
			fmt.Fprintf(w, "    → %s\n", c.Remedy)
		}
	}
	if r.OK {
		fmt.Fprintln(w, "\nAll required checks passed.")
	}
}
