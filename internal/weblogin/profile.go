// Stored login profile: everything the CLI learned about the user at login
// that no API op returns again. Non-secret metadata lives here (0600 JSON
// beside config.yaml); the token itself stays in keychain/config.
// Never edited by regen (hand package).
package weblogin

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/openvoidnet/voidnet-cli/internal/config"
)

// Profile is the persisted post-login picture.
type Profile struct {
	Username   string   `json:"username,omitempty"`
	Scopes     []string `json:"scopes,omitempty"`
	AppIDs     []string `json:"app_ids,omitempty"`
	AllApps    bool     `json:"all_apps"`
	ExpiresAt  string   `json:"expires_at,omitempty"`
	VerifiedAt string   `json:"verified_at,omitempty"`
	TokenID    string   `json:"token_id,omitempty"`
	Last4      string   `json:"last4,omitempty"`
}

func profilePath() (string, error) {
	cfgPath := config.GetConfigPath()
	if cfgPath == "" {
		return "", fmt.Errorf("profile: home directory unknown")
	}
	return filepath.Join(filepath.Dir(cfgPath), "profile.json"), nil
}

// SaveProfile writes the profile with owner-only permissions.
func SaveProfile(p Profile, stderr io.Writer) error {
	path, err := profilePath()
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("profile: encode failed: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("profile: config dir failed: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("profile: write failed: %w", err)
	}
	return nil
}

// LoadProfile reads the stored profile. Missing file means never logged in.
func LoadProfile() (Profile, error) {
	var p Profile
	path, err := profilePath()
	if err != nil {
		return p, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("profile: decode failed: %w", err)
	}
	return p, nil
}

// ClearProfile removes the stored profile. Missing file is success.
func ClearProfile() error {
	path, err := profilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("profile: remove failed: %w", err)
	}
	return nil
}

// ExpiryStatus reports the token lifetime in plain language.
func (p Profile) ExpiryStatus(now time.Time) string {
	if p.ExpiresAt == "" {
		return "expiry unknown"
	}
	exp, err := time.Parse(time.RFC3339, p.ExpiresAt)
	if err != nil {
		return "expiry unreadable"
	}
	d := time.Until(exp)
	if d <= 0 {
		return "expired — run voidnet login"
	}
	days := int(d.Hours() / 24)
	if days == 0 {
		return "expires within 24 hours — run voidnet login soon"
	}
	if days <= 3 {
		return fmt.Sprintf("expires in %d days — run voidnet login soon", days)
	}
	return fmt.Sprintf("valid (%d days left)", days)
}

// BindingLine describes the app binding in plain language.
func (p Profile) BindingLine() string {
	if p.AllApps || len(p.AppIDs) == 0 {
		return "all apps, including future ones"
	}
	if len(p.AppIDs) == 1 {
		return "1 app: " + p.AppIDs[0]
	}
	return fmt.Sprintf("%d apps", len(p.AppIDs))
}
