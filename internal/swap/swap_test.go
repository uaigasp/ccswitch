package swap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/uaigasp/ccswitch/internal/accounts"
)

func writeTestFile(t *testing.T, path string, content map[string]any) {
	t.Helper()
	data, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("marshal fallo: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write fallo: %v", err)
	}
}

func TestReadActive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".credentials.json")

	writeTestFile(t, path, map[string]any{
		"claudeAiOauth": map[string]any{
			"accessToken":  "abc",
			"refreshToken": "def",
		},
		"mcpOAuth": map[string]any{"foo": "bar"},
	})

	got, err := ReadActive(path)
	if err != nil {
		t.Fatalf("read fallo: %v", err)
	}
	if got.AccessToken != "abc" {
		t.Fatalf("got access token %q", got.AccessToken)
	}
}

func TestWriteActivePreservesOtherKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".credentials.json")

	writeTestFile(t, path, map[string]any{
		"claudeAiOauth": map[string]any{"accessToken": "old"},
		"mcpOAuth":      map[string]any{"foo": "bar"},
	})

	newData := accounts.OAuthData{AccessToken: "new-token", RefreshToken: "new-refresh"}
	if err := WriteActive(path, newData); err != nil {
		t.Fatalf("write fallo: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fallo: %v", err)
	}

	var full map[string]json.RawMessage
	if err := json.Unmarshal(raw, &full); err != nil {
		t.Fatalf("unmarshal fallo: %v", err)
	}

	if _, ok := full["mcpOAuth"]; !ok {
		t.Fatal("mcpOAuth se perdio en el write")
	}

	got, err := ReadActive(path)
	if err != nil {
		t.Fatalf("read fallo: %v", err)
	}
	if got.AccessToken != "new-token" {
		t.Fatalf("got access token %q", got.AccessToken)
	}
}
