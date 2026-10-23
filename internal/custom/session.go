package custom

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moderation-api/cli/internal/config"
)

// session is what `modapi login` keeps: an organization key minted for this
// CLI, and the organization it belongs to. The OAuth tokens are not kept; the
// key is revoked on the dashboard's API keys page.
type session struct {
	OrganizationID   string    `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	APIKey           string    `json:"api_key"`
	CreatedAt        time.Time `json:"created_at"`
}

const sessionKeyringKey = "organization-session"

// sessionFile is the fallback when the OS keychain is unavailable or disabled.
func sessionFile() string {
	return filepath.Join(filepath.Dir(config.GetConfigPath()), "session.json")
}

// saveSession stores the session in the OS keychain, or in a 0600 file next
// to the config file. It returns where it was stored.
func saveSession(s session) (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	if config.KeyringAvailable() && config.SetKeyringValue(sessionKeyringKey, string(data)) == nil {
		_ = os.Remove(sessionFile())
		return "OS keychain", nil
	}
	path := sessionFile()
	if err := writePrivateFile(path, data); err != nil {
		return "", fmt.Errorf("could not save the session: %w", err)
	}
	return path, nil
}

// writePrivateFile writes data to a new 0600 file and renames it over path,
// so a file that exists with a looser mode never keeps the secret.
func writePrivateFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".session-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// loadSession returns the stored session, or nil when there is none.
func loadSession() (*session, error) {
	var raw string
	if config.KeyringAvailable() {
		raw = config.GetKeyringValue(sessionKeyringKey)
	}
	if raw == "" {
		data, err := os.ReadFile(sessionFile())
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		raw = string(data)
	}
	var s session
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, fmt.Errorf("the stored session is not valid; run `modapi login` again: %w", err)
	}
	if s.APIKey == "" {
		return nil, nil
	}
	return &s, nil
}

// deleteSession removes the session from both places. It reports whether one
// existed, and whether the keychain was skipped (--no-keyring), in which case
// a session stored there may remain.
func deleteSession() (found, keyringSkipped bool) {
	if config.KeyringAvailable() {
		if config.GetKeyringValue(sessionKeyringKey) != "" {
			found = config.DeleteKeyringValue(sessionKeyringKey) == nil
		}
	} else {
		keyringSkipped = config.KeyringDisabled()
	}
	if err := os.Remove(sessionFile()); err == nil {
		found = true
	}
	return found, keyringSkipped
}
