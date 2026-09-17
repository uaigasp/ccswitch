package swap

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/uaigasp/ccswitch/internal/accounts"
)

func DefaultCredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", ".credentials.json"), nil
}

func ReadActive(path string) (accounts.OAuthData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return accounts.OAuthData{}, err
	}

	var full struct {
		ClaudeAiOauth accounts.OAuthData `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(data, &full); err != nil {
		return accounts.OAuthData{}, err
	}
	return full.ClaudeAiOauth, nil
}

func WriteActive(path string, o accounts.OAuthData) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var full map[string]json.RawMessage
	if err := json.Unmarshal(raw, &full); err != nil {
		return err
	}

	encoded, err := json.Marshal(o)
	if err != nil {
		return err
	}
	full["claudeAiOauth"] = encoded

	out, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
