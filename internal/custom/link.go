package custom

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moderation-api/cli/internal/forms"
	"github.com/moderation-api/cli/internal/output"
	"github.com/moderation-api/cli/internal/sdk/models/components"
	"github.com/moderation-api/cli/internal/sdk/models/operations"
	"github.com/moderation-api/cli/internal/usage"
)

func newLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [project]",
		Short: "Connect this directory to an existing Moderation API project",
		Long: `Connect the project in the current directory to a Moderation API project you
already have.

link makes a new secret key for the project, named after this computer and
directory, and writes it (MODAPI_SECRET_KEY) and the webhook signing secret
(MODAPI_WEBHOOK_SECRET) to the project's env file. The project's other keys
keep working, so you can revoke this one on its own.

Name the project by ID or by name. Without one, link lists your projects to
pick from. It needs an organization key: sign in with "modapi login", or set
MODAPI_ORGANIZATION_KEY.

When the env file already has a key for another project, link asks before
replacing it (use --force in scripts).`,
		Example: `  modapi link
  modapi link "Acme Marketplace"
  modapi link proj_123 --env-file .env.local`,
		Args: cobra.MaximumNArgs(1),
		RunE: runLink,
	}
	f := cmd.Flags()
	f.String("dir", ".", "The project directory")
	f.String("env-file", "", "Env file to write the secrets to (default: .env.local for Next.js, else .env)")
	f.Bool("force", false, "Replace a key for another project without asking")
	f.String("organization-key", "", "Organization key to use. Prefer the MODAPI_ORGANIZATION_KEY env var: a flag value stays in shell history (default: the key from modapi login)")
	return cmd
}

func runLink(cmd *cobra.Command, args []string) error {
	if usage.UsageRequested(cmd) {
		return usage.EmitSchema(cmd, cmd.OutOrStdout())
	}
	dir, err := projectDir(cmd)
	if err != nil {
		return err
	}
	detected := detectStack(dir)
	envFile := chooseEnvFile(cmd, dir, detected)
	existingKey, envFile := existingProjectKey(dir, detected, envFile)
	ref := ""
	if len(args) == 1 {
		ref = strings.TrimSpace(args[0])
	}

	if dryRunPlan(cmd, "link makes no request and writes no files under --dry-run; this is what it would do.", map[string]any{
		"directory":               dir,
		"env_file":                envFile,
		"existing_key_found":      existingKey != "",
		"project":                 ref,
		"organization_key_source": organizationKeySource(cmd),
	}) {
		return nil
	}

	orgKey, err := requireOrganizationKey(cmd, "link needs an organization key to make a key for the project")
	if err != nil {
		return err
	}
	projects, err := listProjects(cmd, orgKey)
	if err != nil {
		return err
	}
	var project *components.Project
	if ref != "" {
		if project, err = findProject(projects, ref); err != nil {
			return err
		}
	} else {
		if len(projects) == 0 {
			return withHints(failure(output.ReasonCLIValidation, "your organization has no projects yet"),
				"Run `modapi init` to create one")
		}
		if !canPrompt(cmd) {
			return chooseProjectError(projects, false)
		}
		id, err := pickProject(cmd, projects, false)
		if err != nil {
			return err
		}
		project, _ = findProject(projects, id)
	}

	result := initResult{EnvFile: relPath(dir, envFile), Stack: detected, VariablesSet: []string{}}
	if existingKey != "" {
		current := currentProjectID(cmd, existingKey)
		if current == project.ID {
			result.Status = "already_set_up"
			result.Project = &initProject{ID: project.ID, Name: project.Name}
			result.DashboardURL = projectDashboardURL(project.ID)
			result.NextSteps = []string{"Check the setup any time: modapi doctor"}
			return render(cmd, result, func(w io.Writer) { printInitResult(w, &result) })
		}
		if err := confirmReplace(cmd, result.EnvFile, project.Name); err != nil {
			return err
		}
	}
	if err := checkEnvFileWritable(dir, envFile, &result); err != nil {
		return err
	}

	if err := linkProject(cmd, dir, envFile, orgKey, project, &result); err != nil {
		return err
	}
	result.NextSteps = nextSteps(detected, &result)
	return render(cmd, result, func(w io.Writer) { printInitResult(w, &result) })
}

func projectDir(cmd *cobra.Command) (string, error) {
	dir, _ := cmd.Flags().GetString("dir")
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", failure(output.ReasonCLIValidation, "invalid --dir: %v", err)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", failure(output.ReasonCLIValidation, "--dir %s is not a directory", dir)
	}
	return dir, nil
}

// requireOrganizationKey returns the organization key, or an error that says
// how to get one.
func requireOrganizationKey(cmd *cobra.Command, why string) (string, error) {
	orgKey, _, err := organizationKey(cmd)
	if err != nil {
		return "", failure(output.ReasonCLIAuthentication, "%v", err)
	}
	if orgKey == "" {
		return "", withHints(failure(output.ReasonCLIAuthentication, "%s", why),
			"Run `modapi login` to sign in in the browser (`modapi login --device` on a machine without a browser)",
			"Or set "+organizationKeyEnv+" to an organization key from the dashboard's API keys page")
	}
	return orgKey, nil
}

// listProjects returns every project the organization key reaches.
func listProjects(cmd *cobra.Command, orgKey string) ([]components.Project, error) {
	api, err := newAPIClient(cmd, orgKey)
	if err != nil {
		return nil, err
	}
	var projects []components.Project
	pageSize := int64(100)
	for page := int64(1); ; page++ {
		res, err := api.Projects.List(cmd.Context(), &operations.ListProjectsRequest{Page: &page, PageSize: &pageSize})
		if err != nil {
			return nil, output.Error(cmd, err)
		}
		if res.Object == nil {
			return projects, nil
		}
		projects = append(projects, res.Object.Data...)
		if !res.Object.Pagination.HasMore {
			return projects, nil
		}
	}
}

// findProject finds a project by ID, or by name regardless of case.
func findProject(projects []components.Project, ref string) (*components.Project, error) {
	for i := range projects {
		if projects[i].ID == ref {
			return &projects[i], nil
		}
	}
	var named []components.Project
	for _, p := range projects {
		if strings.EqualFold(strings.TrimSpace(p.Name), ref) {
			named = append(named, p)
		}
	}
	switch len(named) {
	case 1:
		return &named[0], nil
	case 0:
		return nil, withHints(failure(output.ReasonCLIValidation, "no project with the ID or name %q", ref),
			"See your projects with `modapi projects list`")
	default:
		return nil, withHints(failure(output.ReasonCLIValidation, "more than one project is named %q: %s", ref, describeProjects(named)),
			"Name the project by its ID instead")
	}
}

func describeProjects(projects []components.Project) string {
	const shown = 10
	parts := make([]string, 0, shown)
	for i, p := range projects {
		if i == shown {
			parts = append(parts, fmt.Sprintf("and %d more", len(projects)-shown))
			break
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", p.Name, p.ID))
	}
	return strings.Join(parts, ", ")
}

// chooseProjectError tells an agent or script which projects it can pick.
func chooseProjectError(projects []components.Project, canCreate bool) error {
	hints := []string{"Pass the project to link: --project <id or name>, or `modapi link <id or name>`"}
	if canCreate {
		hints = append(hints, "Or pass --new to create a new project")
	}
	hints = append(hints, "Ask the user which one to use")
	return withHints(failure(output.ReasonCLIValidation, "your organization has %d project(s): %s", len(projects), describeProjects(projects)), hints...)
}

// newProjectChoice is the value of the "Create a new project" option.
const newProjectChoice = "\x00new"

// pickProject asks a person at a terminal which project to link. It returns
// the project ID, or newProjectChoice when canCreate and they pick that.
func pickProject(cmd *cobra.Command, projects []components.Project, canCreate bool) (string, error) {
	options := make([]forms.Option, 0, len(projects)+1)
	if canCreate {
		options = append(options, forms.NewOption("Create a new project", newProjectChoice))
	}
	for _, p := range projects {
		options = append(options, forms.NewOption(fmt.Sprintf("%s (%s)", p.Name, p.ID), p.ID))
	}
	choice := ""
	title := "Link a project"
	if canCreate {
		title = "Link an existing project, or create a new one?"
	}
	field := forms.NewSelect(&choice, options...).Title(title)
	form := forms.New(forms.NewPage("Set up Moderation API", field)).IO(cmd.InOrStdin(), cmd.ErrOrStderr())
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("choose a project: %w", err)
	}
	return choice, nil
}

// currentProjectID is the project an existing key belongs to, or "" when the
// API does not say (an invalid key, or no network).
func currentProjectID(cmd *cobra.Command, key string) string {
	api, err := newQuietAPIClient(cmd, key)
	if err != nil {
		return ""
	}
	res, err := api.Account.Retrieve(cmd.Context(), &operations.GetAccountRequest{})
	if err != nil || res.Object == nil || res.Object.CurrentProject == nil {
		return ""
	}
	return res.Object.CurrentProject.ID
}

func confirmReplace(cmd *cobra.Command, envFile, projectName string) error {
	if force, _ := cmd.Flags().GetBool("force"); force {
		return nil
	}
	if !canPrompt(cmd) {
		return withHints(failure(output.ReasonCLIValidation, "%s in %s is for another project", projectKeyEnv, envFile),
			"Pass --force to replace it with a key for "+projectName)
	}
	replace := false
	field := forms.NewConfirm(&replace).Title(fmt.Sprintf("%s in %s is for another project. Replace it with a key for %s?", projectKeyEnv, envFile, projectName))
	if err := forms.New(forms.NewPage("Link a project", field)).IO(cmd.InOrStdin(), cmd.ErrOrStderr()).Run(); err != nil {
		return fmt.Errorf("link: %w", err)
	}
	if !replace {
		return failure(output.ReasonCLIValidation, "kept the key in %s", envFile)
	}
	return nil
}

// linkProject makes a new key for project and writes it to the env file.
// The caller checks first that the env file is safe to write (checkEnvFileWritable).
func linkProject(cmd *cobra.Command, dir, envFile, orgKey string, project *components.Project, result *initResult) error {
	key, err := createProjectKey(cmd, orgKey, project.ID, linkKeyName(dir))
	if err != nil {
		return err
	}
	if err := writeProjectEnv(cmd, dir, envFile, key, result); err != nil {
		return failure(output.ReasonCLIRuntime, "a key for project %s was made, but writing %s failed: %v. Revoke it on the dashboard's API keys page", project.ID, envFile, err)
	}
	result.Status = "linked"
	result.Project = &initProject{ID: project.ID, Name: project.Name, Domain: deref(project.Domain), Context: deref(project.Context)}
	result.DashboardURL = projectDashboardURL(project.ID)
	return nil
}

// linkKeyName names the key after where it is used: this computer and directory.
func linkKeyName(dir string) string {
	host, _ := os.Hostname()
	host = strings.TrimSuffix(host, ".local")
	name := "CLI · " + filepath.Base(dir)
	if host != "" {
		name += " · " + host
	}
	if r := []rune(name); len(r) > 50 {
		name = string(r[:50])
	}
	return name
}

func projectDashboardURL(projectID string) string {
	return fmt.Sprintf("%s/project/%s/policies", loadSettings().DashboardURL, projectID)
}

// createProjectKey makes a new secret key for a project with the organization key.
func createProjectKey(cmd *cobra.Command, orgKey, projectID, name string) (string, error) {
	// Not the --debug client: it would log the new key.
	api, err := newQuietAPIClient(cmd, orgKey)
	if err != nil {
		return "", err
	}
	stop := output.StartRequestProgress(cmd)
	res, err := api.Projects.APIKeys.Create(cmd.Context(), operations.CreateProjectAPIKeyRequest{
		ID:   projectID,
		Body: operations.CreateProjectAPIKeyRequestBody{Name: &name},
	})
	stop()
	if err != nil {
		return "", output.Error(cmd, err)
	}
	if res.ProjectAPIKey == nil || res.ProjectAPIKey.SecretKey == "" {
		return "", failure(output.ReasonCLIProtocol, "the API made a key for the project but did not return it")
	}
	return res.ProjectAPIKey.SecretKey, nil
}
