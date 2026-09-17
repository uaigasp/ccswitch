package autoswitch

import (
	"testing"
	"time"
)

func TestEvaluateNoSwitchBelowThreshold(t *testing.T) {
	now := time.Now()
	d := Evaluate(50, map[string]float64{"otra": 10}, 90, time.Time{}, now, 300)
	if d.ShouldSwitch {
		t.Fatal("no deberia cambiar por debajo del threshold")
	}
}

func TestEvaluateSwitchesToBestCandidate(t *testing.T) {
	now := time.Now()
	candidates := map[string]float64{"a": 40, "b": 10}
	d := Evaluate(95, candidates, 90, time.Time{}, now, 300)
	if !d.ShouldSwitch {
		t.Fatal("deberia cambiar por encima del threshold")
	}
	if d.TargetAlias != "b" {
		t.Fatalf("esperaba elegir la cuenta con mas margen (b), got %q", d.TargetAlias)
	}
}

func TestEvaluateNoSwitchWithoutCandidates(t *testing.T) {
	now := time.Now()
	d := Evaluate(95, map[string]float64{}, 90, time.Time{}, now, 300)
	if d.ShouldSwitch {
		t.Fatal("no deberia cambiar sin candidatos")
	}
}

func TestEvaluateRespectsCooldown(t *testing.T) {
	now := time.Now()
	lastSwitch := now.Add(-1 * time.Minute)
	d := Evaluate(95, map[string]float64{"a": 10}, 90, lastSwitch, now, 300)
	if d.ShouldSwitch {
		t.Fatal("no deberia cambiar dentro del cooldown")
	}
}
