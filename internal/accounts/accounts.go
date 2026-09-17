package accounts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type OAuthData struct {
	AccessToken           string   `json:"accessToken"`
	RefreshToken          string   `json:"refreshToken"`
	ExpiresAt             int64    `json:"expiresAt"`
	RefreshTokenExpiresAt int64    `json:"refreshTokenExpiresAt"`
	Scopes                []string `json:"scopes"`
	SubscriptionType      string   `json:"subscriptionType"`
	RateLimitTier         string   `json:"rateLimitTier"`
}

type Account struct {
	Alias string    `json:"alias"`
	Email string    `json:"email"`
	OAuth OAuthData `json:"oauth"`
}

type Store struct {
	Active   string    `json:"active"`
	Accounts []Account `json:"accounts"`
}

func (s *Store) Add(a Account) error {
	if _, ok := s.Get(a.Alias); ok {
		return fmt.Errorf("ya existe una cuenta con alias %q", a.Alias)
	}
	s.Accounts = append(s.Accounts, a)
	return nil
}

func (s *Store) Get(alias string) (Account, bool) {
	for _, a := range s.Accounts {
		if a.Alias == alias {
			return a, true
		}
	}
	return Account{}, false
}

func (s *Store) Remove(alias string) error {
	for i, a := range s.Accounts {
		if a.Alias == alias {
			s.Accounts = append(s.Accounts[:i], s.Accounts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no existe una cuenta con alias %q", alias)
}

func Load(dir string) (Store, error) {
	path := filepath.Join(dir, "accounts.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Store{}, nil
	}
	if err != nil {
		return Store{}, err
	}

	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return Store{}, err
	}
	return s, nil
}

func Save(dir string, s Store) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "accounts.json")
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
