package custom

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/moderation-api/cli/internal/client"
	"github.com/moderation-api/cli/internal/config"
)

// commandsWithoutAPIKey are the top-level commands that do not call the API
// with the secret key, or that find their own key.
var commandsWithoutAPIKey = map[string]bool{
	"login": true, "logout": true, "init": true, "link": true, "doctor": true,
	"auth": true, "configure": true, "whoami": true,
	"help": true, "completion": true, "version": true,
}

// managementCommands call the management API, which takes an organization key.
var managementCommands = map[string]bool{"projects": true, "channels": true}

// useFallbackKey runs before every command. When an API command has no
// secret key (flag, MODAPI_SECRET_KEY, `modapi configure`), it uses the key
// `modapi init` wrote to the project's env file in the current directory, or
// the organization key from `modapi login` / MODAPI_ORGANIZATION_KEY. The
// management commands try the organization key first. The key reaches the
// request through MODAPI_SECRET_KEY, which the generated client reads.
func useFallbackKey(cmd *cobra.Command) error {
	top := topLevelName(cmd)
	if top == "" || commandsWithoutAPIKey[top] {
		return nil
	}
	if key, _ := config.ResolveRequestSecurityCredential(cmd, "secret-key"); key != "" {
		return nil
	}

	key, err := fallbackKey(managementCommands[top])
	if err != nil {
		return err
	}
	if key == "" {
		if client.IsDryRun(cmd) {
			return nil
		}
		return withHints(
			failure("CLI_AUTHENTICATION", "no API key: this command needs a project key or an organization key"),
			"Run `modapi init` in your project's directory to create or link a project and write its key to the env file",
			"Or run `modapi login` to use your organization key",
			"Or set MODAPI_SECRET_KEY, or pass --secret-key",
		)
	}
	return os.Setenv(projectKeyEnv, key)
}

func fallbackKey(organizationFirst bool) (string, error) {
	projectKey := func() (string, error) {
		dir, err := os.Getwd()
		if err != nil {
			return "", nil
		}
		v, _ := findEnvValue(dir, detectStack(dir), projectKeyEnv)
		return v, nil
	}
	orgKey := func() (string, error) {
		if v := os.Getenv(organizationKeyEnv); v != "" {
			return v, nil
		}
		s, err := loadSession()
		if err != nil || s == nil {
			return "", err
		}
		return s.APIKey, nil
	}

	order := []func() (string, error){projectKey, orgKey}
	if organizationFirst {
		order = []func() (string, error){orgKey, projectKey}
	}
	for _, find := range order {
		key, err := find()
		if err != nil {
			return "", err
		}
		if key != "" {
			return key, nil
		}
	}
	return "", nil
}

// topLevelName is the name of the command right under the root.
func topLevelName(cmd *cobra.Command) string {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Parent() != nil && c.Parent().Parent() == nil {
			return c.Name()
		}
	}
	return ""
}
