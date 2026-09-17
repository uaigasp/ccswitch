package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReturnsDefaultWhenFileMissing(t *testing.T) {
	dir := t.TempDir()

	c, err := Load(dir)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	want := Default()
	if c != want {
		t.Fatalf("got %+v, want %+v", c, want)
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()

	c := Config{ThresholdPercent: 80, IntervalSeconds: 30, CooldownSeconds: 120}
	if err := Save(dir, c); err != nil {
		t.Fatalf("save fallo: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("load fallo: %v", err)
	}
	if got != c {
		t.Fatalf("got %+v, want %+v", got, c)
	}

	if _, err := os.Stat(filepath.Join(dir, "config.json")); err != nil {
		t.Fatalf("no se creo config.json: %v", err)
	}
}

func TestLoadFillsDefaultsForMissingFields(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"thresholdPercent": 80}`), 0o644); err != nil {
		t.Fatalf("no se pudo escribir config.json: %v", err)
	}

	c, err := Load(dir)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	want := Default()
	want.ThresholdPercent = 80
	if c != want {
		t.Fatalf("got %+v, want %+v", c, want)
	}
}
