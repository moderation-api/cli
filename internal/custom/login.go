package custom

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/moderation-api/cli/internal/output"
	"github.com/moderation-api/cli/internal/usage"
)

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in with your Moderation API account",
		Long: `Sign in with your Moderation API account in the browser.

The CLI gets an organization key for the organization you pick and keeps it in
the OS keychain (or a 0600 file when there is no keychain). modapi init uses it
to create projects. Revoke it on the dashboard's API keys page.

On a machine without a browser, such as over SSH, use --device: you confirm a
code on another device. In CI, skip login and set MODAPI_ORGANIZATION_KEY.

This is not the same as "modapi auth login", which stores a project key you
already have.`,
		Example: `  modapi login
  modapi login --device`,
		Args: cobra.NoArgs,
		RunE: runLogin,
	}
	cmd.Flags().Bool("device", false, "Sign in by confirming a code on another device (for SSH and machines without a browser)")
	cmd.Flags().Bool("no-browser", false, "Print the sign-in URL instead of opening the browser")
	cmd.Flags().Duration("wait", 5*time.Minute, "How long to wait for you to finish signing in")
	return cmd
}

type loginResult struct {
	Organization struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"organization"`
	StoredIn string `json:"stored_in"`
	KeyHint  string `json:"key_hint"`
}

func runLogin(cmd *cobra.Command, _ []string) error {
	if usage.UsageRequested(cmd) {
		return usage.EmitSchema(cmd, cmd.OutOrStdout())
	}
	if dryRunPlan(cmd, "login signs in in the browser and stores an organization key locally; nothing was changed.", nil) {
		return nil
	}
	cfg := loadSettings()
	if cfg.ClientID == "" {
		return withHints(failure(output.ReasonCLIUnavailable, "sign-in is not set up in this build of modapi (no OAuth client ID)"),
			"Set "+organizationKeyEnv+" to an organization key from the dashboard's API keys page",
			"Or update modapi")
	}

	device, _ := cmd.Flags().GetBool("device")
	noBrowser, _ := cmd.Flags().GetBool("no-browser")
	wait, _ := cmd.Flags().GetDuration("wait")
	// Sign-in prompts go to stderr so stdout stays machine-readable.
	prompt := cmd.ErrOrStderr()
	httpClient := &http.Client{Timeout: 30 * time.Second}
	ctx := cmd.Context()

	server, err := discover(ctx, httpClient, cfg.Issuer)
	if err != nil {
		return failure(output.ReasonCLIConnection, "%v", err)
	}

	var tokens *tokenSet
	if device || (!noBrowser && !hasBrowser()) {
		tokens, err = deviceLogin(ctx, deviceLoginOptions{
			Server: server, ClientID: cfg.ClientID, Resource: cfg.Resource,
			Scopes: loginScopes, Prompt: prompt, HTTP: httpClient,
		})
	} else {
		var open func(string) error
		if !noBrowser {
			open = openBrowser
		}
		tokens, err = browserLogin(ctx, browserLoginOptions{
			Server: server, ClientID: cfg.ClientID, Resource: cfg.Resource,
			Scopes: loginScopes, Open: open, Prompt: prompt, Wait: wait, HTTP: httpClient,
		})
	}
	if err != nil {
		return withHints(failure(output.ReasonCLIAuthentication, "sign-in failed: %v", err),
			"Run `modapi login` again",
			"Without a browser, use `modapi login --device`")
	}

	hostname, _ := os.Hostname()
	exchanged, err := exchangeForOrganizationKey(ctx, httpClient, cfg.KeyExchangeURL, tokens.AccessToken, hostname)
	if errors.Is(err, errExchangeNotAvailable) {
		return withHints(failure(output.ReasonCLIUnavailable, "%v", err),
			"Set "+organizationKeyEnv+" to an organization key from the dashboard's API keys page")
	}
	if err != nil {
		return failure(output.ReasonCLIAuthentication, "%v", err)
	}

	storedIn, err := saveSession(session{
		OrganizationID:   exchanged.OrganizationID,
		OrganizationName: exchanged.OrganizationName,
		APIKey:           exchanged.APIKey,
		CreatedAt:        time.Now().UTC(),
	})
	if err != nil {
		return failure(output.ReasonCLIRuntime, "%v", err)
	}

	var result loginResult
	result.Organization.ID = exchanged.OrganizationID
	result.Organization.Name = exchanged.OrganizationName
	result.StoredIn = storedIn
	result.KeyHint = maskKey(exchanged.APIKey)
	return render(cmd, result, func(w io.Writer) {
		fmt.Fprintf(w, "Signed in to %s.\n", exchanged.OrganizationName)
		fmt.Fprintf(w, "Organization key %s stored in %s.\n", result.KeyHint, storedIn)
		fmt.Fprintln(w, "\nNext: run `modapi init` in your project.")
	})
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the organization key stored by modapi login",
		Long: `Remove the organization key that modapi login stored on this machine.

The key itself stays valid until you delete it on the dashboard's API keys page.
Project keys stored with "modapi auth login" or "modapi configure" are not touched.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if usage.UsageRequested(cmd) {
				return usage.EmitSchema(cmd, cmd.OutOrStdout())
			}
			if dryRunPlan(cmd, "logout removes the locally stored organization key; nothing was changed.", nil) {
				return nil
			}
			removed, keyringSkipped := deleteSession()
			apiKeysURL := loadSettings().DashboardURL + "/settings/api-keys"
			result := map[string]any{"removed": removed, "keychain_skipped": keyringSkipped, "revoke_url": apiKeysURL}
			return render(cmd, result, func(w io.Writer) {
				if keyringSkipped {
					fmt.Fprintln(w, "The OS keychain was skipped (--no-keyring); a session stored there was not removed.")
				}
				if !removed {
					fmt.Fprintln(w, "No session found to remove.")
					return
				}
				fmt.Fprintln(w, "Signed out. The organization key was removed from this machine.")
				fmt.Fprintf(w, "To revoke the key itself, delete it at %s.\n", apiKeysURL)
			})
		},
	}
}
