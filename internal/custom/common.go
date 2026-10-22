package custom

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/moderation-api/cli/internal/client"
	"github.com/moderation-api/cli/internal/flagutil"
	"github.com/moderation-api/cli/internal/output"
	"github.com/moderation-api/cli/internal/sdk"
	"github.com/moderation-api/cli/internal/sdk/models/components"
	"github.com/moderation-api/cli/internal/testclient"
)

// Register adds the hand-written commands to the root command.
func Register(root *cobra.Command) {
	root.AddCommand(newLoginCmd(), newLogoutCmd(), newInitCmd(), newDoctorCmd())
}

// failure is an error the CLI's error renderer classifies by reason:
// CLI_VALIDATION exits 2, CLI_AUTHENTICATION exits 3, the others exit 1.
func failure(reason, format string, args ...any) error {
	return output.WithCLIReason(fmt.Errorf(format, args...), reason)
}

// hintedError replaces the generic hints of the error's type with hints that
// fit the command. The error renderer finds these methods by interface.
type hintedError struct {
	error
	hints []string
}

func (e hintedError) CLIHints() []string     { return e.hints }
func (e hintedError) CLIOmitTypeHints() bool { return true }
func (e hintedError) Unwrap() error          { return e.error }

func withHints(err error, hints ...string) error {
	return hintedError{error: err, hints: hints}
}

// newAPIClient makes an SDK client that calls with key and honours the
// global --server-url, --timeout, --debug and --dry-run flags.
func newAPIClient(cmd *cobra.Command, key string) (*sdk.ModerationAPI, error) {
	opts := []sdk.SDKOption{sdk.WithSecurity(components.Security{SecretKey: key})}
	if serverURL, _ := flagutil.GetStringFlag(cmd, "server-url"); serverURL != "" {
		if err := flagutil.ValidateServerURL(serverURL); err != nil {
			return nil, err
		}
		opts = append(opts, sdk.WithServerURL(serverURL))
	}
	if timeout, _ := flagutil.GetStringFlag(cmd, "timeout"); timeout != "" {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			return nil, flagutil.WithCLIValidation(fmt.Errorf("invalid --timeout value %q: %w", timeout, err))
		}
		opts = append(opts, sdk.WithTimeout(d))
	}
	var httpClient client.HTTPClient = &http.Client{}
	if test := testclient.NewTestHTTPClient(); test != nil {
		httpClient = test
	}
	opts = append(opts, sdk.WithClient(client.WrapClientForDiagnostics(cmd, httpClient)))
	return sdk.New(opts...), nil
}

// render writes value as structured output in machine and agent mode, and
// calls human otherwise.
func render(cmd *cobra.Command, value any, human func(w io.Writer)) error {
	if output.IsMachineMode(cmd) {
		// The output layer unwraps structs as SDK responses (it prints their
		// first field), so hand it the value as a plain map.
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		var plain map[string]any
		if err := json.Unmarshal(data, &plain); err != nil {
			return err
		}
		return output.LocalResult(cmd, plain)
	}
	human(cmd.OutOrStdout())
	return nil
}

// dryRunPlan reports what a command would do under --dry-run, without
// network calls, keychain reads or file writes. It returns true when it did.
func dryRunPlan(cmd *cobra.Command, message string, plan any) bool {
	if !client.IsDryRun(cmd) {
		return false
	}
	if client.IsJSONDryRun(cmd) || output.IsAgentMode() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetEscapeHTML(false)
		_ = enc.Encode(map[string]any{
			"dry_run": true,
			"local":   true,
			"command": cmd.CommandPath(),
			"message": message,
			"plan":    plan,
		})
		return true
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "[DRY-RUN] "+message)
	if plan != nil {
		data, _ := json.MarshalIndent(plan, "", "  ")
		fmt.Fprintln(cmd.ErrOrStderr(), string(data))
	}
	return true
}

// canPrompt is true for a person at a terminal who has not turned prompts off.
func canPrompt(cmd *cobra.Command) bool {
	if output.IsMachineMode(cmd) {
		return false
	}
	if off, _ := flagutil.GetBoolFlag(cmd, "no-interactive"); off {
		return false
	}
	if on, changed := flagutil.GetBoolFlag(cmd, "interactive"); changed && !on {
		return false
	}
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// openBrowser opens url in the default browser.
func openBrowser(url string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		c = exec.Command("xdg-open", url)
	}
	return c.Start()
}

// hasBrowser guesses whether this machine can open a browser for sign-in.
func hasBrowser() bool {
	if os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" {
		return false
	}
	if runtime.GOOS == "linux" {
		return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	}
	return true
}

// maskKey shows the start and end of a key so people can tell keys apart.
func maskKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("•", len(key))
	}
	return key[:7] + "…" + key[len(key)-4:]
}
