package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ThresholdPercent int `json:"thresholdPercent"`
	IntervalSeconds  int `json:"intervalSeconds"`
	CooldownSeconds  int `json:"cooldownSeconds"`
}

func Default() Config {
	return Config{
		ThresholdPercent: 90,
		IntervalSeconds:  60,
		CooldownSeconds:  300,
	}
}

func Load(dir string) (Config, error) {
	path := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}

	c := Default()
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, err
	}
	return c, nil
}

func Save(dir string, c Config) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "config.json")
	return os.WriteFile(path, data, 0o644)
}
