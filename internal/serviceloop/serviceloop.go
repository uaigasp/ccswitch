package serviceloop

import (
	"fmt"
	"time"

	"github.com/uaigasp/ccswitch/internal/accounts"
	"github.com/uaigasp/ccswitch/internal/autoswitch"
	"github.com/uaigasp/ccswitch/internal/config"
	"github.com/uaigasp/ccswitch/internal/swap"
	"github.com/uaigasp/ccswitch/internal/usage"
)

type usageFetcher interface {
	Fetch(accessToken string) (float64, error)
}

type realUsageFetcher struct {
	client *usage.Client
}

func (f realUsageFetcher) Fetch(accessToken string) (float64, error) {
	result, err := f.client.Fetch(accessToken)
	if err != nil {
		return 0, err
	}
	return result.FiveHour.UtilizationPercent, nil
}

func RunOnce(dir, credPath string, cfg config.Config, lastSwitch time.Time, dryRun bool) (autoswitch.Decision, error) {
	fetcher := realUsageFetcher{client: usage.NewClient()}
	return doRunOnce(dir, credPath, fetcher, cfg, lastSwitch, dryRun)
}

func doRunOnce(dir, credPath string, fetcher usageFetcher, cfg config.Config, lastSwitch time.Time, dryRun bool) (autoswitch.Decision, error) {
	decision, store, err := runOnceWithFetcher(dir, fetcher, cfg, lastSwitch, dryRun)
	if err != nil {
		return autoswitch.Decision{}, err
	}

	if decision.ShouldSwitch && !dryRun {
		target, ok := store.Get(decision.TargetAlias)
		if !ok {
			return decision, fmt.Errorf("la cuenta %q ya no existe en el store", decision.TargetAlias)
		}

		if store.Active != "" {
			if live, err := swap.ReadActive(credPath); err == nil {
				for i := range store.Accounts {
					if store.Accounts[i].Alias == store.Active {
						store.Accounts[i].OAuth = live
						break
					}
				}
			}
		}

		if err := swap.WriteActive(credPath, target.OAuth); err != nil {
			return decision, err
		}
		store.Active = decision.TargetAlias
		if err := accounts.Save(dir, store); err != nil {
			return decision, err
		}
	}

	return decision, nil
}

func runOnceWithFetcher(dir string, fetcher usageFetcher, cfg config.Config, lastSwitch time.Time, dryRun bool) (autoswitch.Decision, accounts.Store, error) {
	store, err := accounts.Load(dir)
	if err != nil {
		return autoswitch.Decision{}, accounts.Store{}, err
	}

	if store.Active == "" || len(store.Accounts) < 2 {
		return autoswitch.Decision{}, store, nil
	}

	active, ok := store.Get(store.Active)
	if !ok {
		return autoswitch.Decision{}, store, nil
	}

	activePct, err := fetcher.Fetch(active.OAuth.AccessToken)
	if err != nil {
		return autoswitch.Decision{}, accounts.Store{}, err
	}

	now := time.Now()
	if now.Sub(lastSwitch) < time.Duration(cfg.CooldownSeconds)*time.Second {
		return autoswitch.Decision{}, store, nil
	}
	if activePct < float64(cfg.ThresholdPercent) {
		return autoswitch.Decision{}, store, nil
	}

	candidates := map[string]float64{}
	for _, a := range store.Accounts {
		if a.Alias == store.Active {
			continue
		}
		pct, err := fetcher.Fetch(a.OAuth.AccessToken)
		if err != nil {
			continue
		}
		candidates[a.Alias] = pct
	}

	return autoswitch.Evaluate(activePct, candidates, cfg.ThresholdPercent, lastSwitch, now, cfg.CooldownSeconds), store, nil
}
