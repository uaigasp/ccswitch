package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/uaigasp/ccswitch/internal/accounts"
	"github.com/uaigasp/ccswitch/internal/swap"
	"github.com/uaigasp/ccswitch/internal/usage"
)

func appDataDir() (string, error) {
	base := os.Getenv("APPDATA")
	if base == "" {
		return "", fmt.Errorf("no se encontro la variable de entorno APPDATA")
	}
	return filepath.Join(base, "ccswitch"), nil
}

func runAdd(alias, email string) error {
	if alias == "" {
		return fmt.Errorf("hace falta --alias")
	}

	credPath, err := swap.DefaultCredentialsPath()
	if err != nil {
		return err
	}

	oauth, err := swap.ReadActive(credPath)
	if err != nil {
		return fmt.Errorf("no se pudo leer la sesion activa de claude code: %w", err)
	}

	dir, err := appDataDir()
	if err != nil {
		return err
	}

	store, err := accounts.Load(dir)
	if err != nil {
		return err
	}

	if err := store.Add(accounts.Account{Alias: alias, Email: email, OAuth: oauth}); err != nil {
		return err
	}
	if store.Active == "" {
		store.Active = alias
	}

	return accounts.Save(dir, store)
}

func runList() error {
	dir, err := appDataDir()
	if err != nil {
		return err
	}

	store, err := accounts.Load(dir)
	if err != nil {
		return err
	}

	if len(store.Accounts) == 0 {
		fmt.Println("no hay cuentas guardadas todavia. usa: ccswitch add --alias <nombre>")
		return nil
	}

	client := usage.NewClient()
	for _, a := range store.Accounts {
		marker := " "
		if a.Alias == store.Active {
			marker = "*"
		}

		result, err := client.Fetch(a.OAuth.AccessToken)
		if err != nil {
			fmt.Printf("%s %s (%s) - no se pudo consultar el uso: %v\n", marker, a.Alias, a.Email, err)
			continue
		}
		fmt.Printf("%s %s (%s) - 5h: %.0f%%  7d: %.0f%%\n", marker, a.Alias, a.Email, result.FiveHour.UtilizationPercent, result.SevenDay.UtilizationPercent)
	}
	return nil
}

func runStatus() error {
	dir, err := appDataDir()
	if err != nil {
		return err
	}

	store, err := accounts.Load(dir)
	if err != nil {
		return err
	}

	if store.Active == "" {
		fmt.Println("no hay ninguna cuenta activa todavia")
		return nil
	}

	active, ok := store.Get(store.Active)
	if !ok {
		return fmt.Errorf("la cuenta activa %q no esta en el store", store.Active)
	}

	client := usage.NewClient()
	result, err := client.Fetch(active.OAuth.AccessToken)
	if err != nil {
		return fmt.Errorf("no se pudo consultar el uso: %w", err)
	}

	fmt.Printf("cuenta activa: %s (%s)\n5h: %.0f%%\n7d: %.0f%%\n", active.Alias, active.Email, result.FiveHour.UtilizationPercent, result.SevenDay.UtilizationPercent)
	return nil
}

func runSwitch(alias string) error {
	if alias == "" {
		return fmt.Errorf("hace falta indicar el alias. uso: ccswitch switch <alias>")
	}

	dir, err := appDataDir()
	if err != nil {
		return err
	}

	store, err := accounts.Load(dir)
	if err != nil {
		return err
	}

	target, ok := store.Get(alias)
	if !ok {
		return fmt.Errorf("no existe una cuenta con alias %q", alias)
	}

	credPath, err := swap.DefaultCredentialsPath()
	if err != nil {
		return err
	}

	if err := swap.WriteActive(credPath, target.OAuth); err != nil {
		return fmt.Errorf("no se pudo escribir las credenciales: %w", err)
	}

	store.Active = alias
	return accounts.Save(dir, store)
}

func parseAddFlags(args []string) (alias, email string) {
	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--alias":
			alias = args[i+1]
		case "--email":
			email = args[i+1]
		}
	}
	return alias, email
}
