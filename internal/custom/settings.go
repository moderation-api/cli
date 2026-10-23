// Package custom holds hand-written commands that sit next to the generated
// ones: `modapi login`, `logout`, `init` and `doctor`. Nothing in this package
// is generated, so regeneration leaves it alone; it is wired in with one call
// in internal/cli/root.go (kept by Speakeasy persistent edits).
package custom

import (
	"os"
	"strings"
)

// OAuthClientID is the Clerk OAuth application the CLI signs in with: a public
// client with PKCE, the redirect URI http://127.0.0.1/callback and the device
// authorization grant on ("Moderation API CLI" on the production Clerk
// instance). It is not a secret. Override it at build time with
// -ldflags "-X github.com/moderation-api/cli/internal/custom.OAuthClientID=...",
// or at run time with MODAPI_OAUTH_CLIENT_ID.
var OAuthClientID = "4Cis1Jh7mS6MINJg"

const (
	defaultIssuer       = "https://clerk.moderationapi.com"
	defaultDashboardURL = "https://dash.moderationapi.com"
	// defaultResource is the audience the CLI asks Clerk to put in its access
	// token; the dashboard's CLI key exchange accepts only this audience.
	defaultResource = "https://dash.moderationapi.com/api/cli/key"

	// projectKeyEnv and webhookSecretEnv are the variables every SDK reads.
	projectKeyEnv    = "MODAPI_SECRET_KEY"
	webhookSecretEnv = "MODAPI_WEBHOOK_SECRET"
	// organizationKeyEnv lets CI and agents without a browser run init with an organization key.
	organizationKeyEnv = "MODAPI_ORGANIZATION_KEY"
)

// loginScopes: offline_access is not needed because the CLI keeps no tokens;
// user:org:read makes Clerk show an organization picker and put org_id in the token.
var loginScopes = []string{"openid", "profile", "email", "user:org:read"}

type settings struct {
	Issuer       string
	ClientID     string
	Resource     string
	DashboardURL string
}

// loadSettings reads the sign-in settings. The env vars exist for staging and tests.
func loadSettings() settings {
	return settings{
		Issuer:       envOr("MODAPI_OAUTH_ISSUER", defaultIssuer),
		ClientID:     envOr("MODAPI_OAUTH_CLIENT_ID", OAuthClientID),
		Resource:     envOr("MODAPI_OAUTH_RESOURCE", defaultResource),
		DashboardURL: strings.TrimSuffix(envOr("MODAPI_DASHBOARD_URL", defaultDashboardURL), "/"),
	}
}

func envOr(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}
