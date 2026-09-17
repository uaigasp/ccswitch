package serviceloop

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/uaigasp/ccswitch/internal/accounts"
	"github.com/uaigasp/ccswitch/internal/config"
)

type fakeUsageFetcher struct {
	byToken map[string]float64
}

func (f fakeUsageFetcher) Fetch(accessToken string) (float64, error) {
	return f.byToken[accessToken], nil
}

func TestRunOnceDryRunDoesNotWriteCredentials(t *testing.T) {
	dir := t.TempDir()

	store := accounts.Store{
		Active: "principal",
		Accounts: []accounts.Account{
			{Alias: "principal", OAuth: accounts.OAuthData{AccessToken: "tok-a"}},
			{Alias: "secundaria", OAuth: accounts.OAuthData{AccessToken: "tok-b"}},
		},
	}
	accounts.Save(dir, store)

	fetcher := fakeUsageFetcher{byToken: map[string]float64{"tok-a": 95, "tok-b": 10}}
	cfg := config.Config{ThresholdPercent: 90, CooldownSeconds: 300}

	decision, _, err := runOnceWithFetcher(dir, fetcher, cfg, time.Time{}, true)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !decision.ShouldSwitch || decision.TargetAlias != "secundaria" {
		t.Fatalf("got decision %+v", decision)
	}

	after, _ := accounts.Load(dir)
	if after.Active != "principal" {
		t.Fatalf("en dry run el active no deberia cambiar, got %q", after.Active)
	}
}

func TestRunOnceAppliesSwapWhenNotDryRun(t *testing.T) {
	dir := t.TempDir()
	credPath := t.TempDir() + "/credentials.json"

	// Create a fake credentials file
	credentialsContent := map[string]interface{}{
		"claudeAiOauth": map[string]string{
			"accessToken": "tok-a",
		},
	}
	credentialsData, err := json.Marshal(credentialsContent)
	if err != nil {
		t.Fatalf("error marshaling credentials: %v", err)
	}
	if err := os.WriteFile(credPath, credentialsData, 0o600); err != nil {
		t.Fatalf("error writing credentials file: %v", err)
	}

	store := accounts.Store{
		Active: "principal",
		Accounts: []accounts.Account{
			{Alias: "principal", OAuth: accounts.OAuthData{AccessToken: "tok-a"}},
			{Alias: "secundaria", OAuth: accounts.OAuthData{AccessToken: "tok-b"}},
		},
	}
	accounts.Save(dir, store)

	fetcher := fakeUsageFetcher{byToken: map[string]float64{"tok-a": 95, "tok-b": 10}}
	cfg := config.Config{ThresholdPercent: 90, CooldownSeconds: 300}

	decision, err := doRunOnce(dir, credPath, fetcher, cfg, time.Time{}, false)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !decision.ShouldSwitch || decision.TargetAlias != "secundaria" {
		t.Fatalf("got decision %+v", decision)
	}

	// Verify Active was changed in store
	after, _ := accounts.Load(dir)
	if after.Active != "secundaria" {
		t.Fatalf("despues del swap el active deberia ser secundaria, got %q", after.Active)
	}

	// Verify credentials file was updated
	credData, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("error reading credentials file: %v", err)
	}
	var credPayload map[string]interface{}
	if err := json.Unmarshal(credData, &credPayload); err != nil {
		t.Fatalf("error unmarshaling credentials: %v", err)
	}
	claudeAiOauth, ok := credPayload["claudeAiOauth"].(map[string]interface{})
	if !ok {
		t.Fatalf("claudeAiOauth not found or not a map")
	}
	accessToken, ok := claudeAiOauth["accessToken"].(string)
	if !ok {
		t.Fatalf("accessToken not found or not a string")
	}
	if accessToken != "tok-b" {
		t.Fatalf("credentials should have tok-b, got %q", accessToken)
	}
}
