package serviceloop

import (
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

	decision, err := runOnceWithFetcher(dir, fetcher, cfg, time.Time{}, true)
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
