package autoswitch

import "time"

type Decision struct {
	ShouldSwitch bool
	TargetAlias  string
}

func Evaluate(
	activeFiveHourPct float64,
	candidates map[string]float64,
	thresholdPercent int,
	lastSwitch time.Time,
	now time.Time,
	cooldownSeconds int,
) Decision {
	if now.Sub(lastSwitch) < time.Duration(cooldownSeconds)*time.Second {
		return Decision{}
	}

	if activeFiveHourPct < float64(thresholdPercent) {
		return Decision{}
	}

	bestAlias := ""
	bestPct := 0.0
	found := false
	for alias, pct := range candidates {
		if !found || pct < bestPct {
			bestAlias = alias
			bestPct = pct
			found = true
		}
	}

	if !found {
		return Decision{}
	}

	return Decision{ShouldSwitch: true, TargetAlias: bestAlias}
}
