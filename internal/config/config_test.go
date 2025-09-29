package config

import (
	"os"
	"testing"
)

func TestLoadReadsEnvAndBuildsDSN(t *testing.T) {
	_ = os.Setenv("DB_HOST", "dbhost")
	_ = os.Setenv("DB_PORT", "1234")
	_ = os.Setenv("DB_USER", "u")
	_ = os.Setenv("DB_PASSWORD", "p")
	_ = os.Setenv("DB_NAME", "n")
	_ = os.Setenv("SERVER_HOST", "127.0.0.1")
	_ = os.Setenv("SERVER_PORT", "9090")
	_ = os.Setenv("APP_ENV", "test")
	_ = os.Setenv("LOG_LEVEL", "debug")
	defer func() { os.Clearenv() }()

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Host != "dbhost" || cfg.Database.Port != 1234 || cfg.Server.Port != "9090" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if cfg.Database.DSN == "" {
		t.Fatal("dsn should be built")
	}
}

func TestGetEnvHelpers(t *testing.T) {
	_ = os.Unsetenv("X")
	if v := getEnv("X", "d"); v != "d" {
		t.Fatal(v)
	}
	_ = os.Setenv("Y", "42")
	if v := getEnvAsInt("Y", 0); v != 42 {
		t.Fatal(v)
	}
}

func TestGetEnvAsInt_InvalidValue(t *testing.T) {
	_ = os.Setenv("INVALID_INT", "not-a-number")
	defer func() { _ = os.Unsetenv("INVALID_INT") }()

	if v := getEnvAsInt("INVALID_INT", 100); v != 100 {
		t.Errorf("expected 100 (default), got %d", v)
	}
}
