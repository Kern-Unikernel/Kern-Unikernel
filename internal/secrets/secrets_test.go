package secrets

import (
	"os"
	"testing"
)

func TestSetAndGetSecret(t *testing.T) {
	// Usar directorio temporal para los tests
	home := t.TempDir()
	os.Setenv("HOME", home)

	if err := SetSecret("db_url", "postgres://user:pass@localhost/db"); err != nil {
		t.Fatalf("error setting secret: %v", err)
	}

	value, err := GetSecret("db_url")
	if err != nil {
		t.Fatalf("error getting secret: %v", err)
	}

	if value != "postgres://user:pass@localhost/db" {
		t.Errorf("expected postgres://user:pass@localhost/db, got %s", value)
	}
}

func TestGetNonExistentSecret(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)

	_, err := GetSecret("nonexistent")
	if err == nil {
		t.Error("should return error for nonexistent secret")
	}
}

func TestSetSecretOverwrite(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)

	if err := SetSecret("key", "value1"); err != nil {
		t.Fatal(err)
	}
	if err := SetSecret("key", "value2"); err != nil {
		t.Fatal(err)
	}

	value, err := GetSecret("key")
	if err != nil {
		t.Fatal(err)
	}
	if value != "value2" {
		t.Errorf("expected value2, got %s", value)
	}
}

func TestResolveSecrets(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)

	if err := SetSecret("db_url", "postgres://localhost/db"); err != nil {
		t.Fatal(err)
	}

	envMap := map[string]string{
		"DATABASE_URL": "secret:db_url",
		"NODE_ENV":     "production",
	}

	resolved, err := ResolveSecrets(envMap)
	if err != nil {
		t.Fatalf("error resolving secrets: %v", err)
	}

	if resolved["DATABASE_URL"] != "postgres://localhost/db" {
		t.Errorf("expected postgres://localhost/db, got %s", resolved["DATABASE_URL"])
	}
	if resolved["NODE_ENV"] != "production" {
		t.Errorf("expected production, got %s", resolved["NODE_ENV"])
	}
}

func TestResolveSecretsNonExistent(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)

	envMap := map[string]string{
		"DATABASE_URL": "secret:nonexistent",
	}

	_, err := ResolveSecrets(envMap)
	if err == nil {
		t.Error("should return error for nonexistent secret")
	}
}
