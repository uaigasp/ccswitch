package accounts

import "testing"

func TestStoreAddAndGet(t *testing.T) {
	var s Store

	a := Account{Alias: "personal", Email: "a@example.com"}
	if err := s.Add(a); err != nil {
		t.Fatalf("add fallo: %v", err)
	}

	got, ok := s.Get("personal")
	if !ok {
		t.Fatal("esperaba encontrar la cuenta")
	}
	if got.Email != "a@example.com" {
		t.Fatalf("got email %q", got.Email)
	}
}

func TestStoreAddDuplicateAliasFails(t *testing.T) {
	var s Store
	s.Add(Account{Alias: "personal"})

	err := s.Add(Account{Alias: "personal"})
	if err == nil {
		t.Fatal("esperaba error por alias duplicado")
	}
}

func TestStoreRemove(t *testing.T) {
	var s Store
	s.Add(Account{Alias: "personal"})

	if err := s.Remove("personal"); err != nil {
		t.Fatalf("remove fallo: %v", err)
	}

	if _, ok := s.Get("personal"); ok {
		t.Fatal("la cuenta no deberia estar mas")
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()

	var s Store
	s.Add(Account{Alias: "personal", Email: "a@example.com"})
	s.Active = "personal"

	if err := Save(dir, s); err != nil {
		t.Fatalf("save fallo: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("load fallo: %v", err)
	}

	if got.Active != "personal" {
		t.Fatalf("got active %q", got.Active)
	}
	if len(got.Accounts) != 1 || got.Accounts[0].Email != "a@example.com" {
		t.Fatalf("got accounts %+v", got.Accounts)
	}
}

func TestLoadReturnsEmptyStoreWhenFileMissing(t *testing.T) {
	dir := t.TempDir()

	s, err := Load(dir)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(s.Accounts) != 0 {
		t.Fatalf("esperaba store vacio, got %+v", s)
	}
}
