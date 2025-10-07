package config

import (
	"testing"

	"github.com/robbiew/retrograde/internal/database"
)

func TestGetDefaultConfigAuthDefaults(t *testing.T) {
	cfg := GetDefaultConfig()
	if cfg.Configuration.Auth.MaxFailedAttempts != 5 {
		t.Fatalf("expected MaxFailedAttempts default 5, got %d", cfg.Configuration.Auth.MaxFailedAttempts)
	}
	if cfg.Configuration.Auth.AccountLockMinutes != 15 {
		t.Fatalf("expected AccountLockMinutes default 15, got %d", cfg.Configuration.Auth.AccountLockMinutes)
	}
	if cfg.Configuration.Auth.PasswordAlgorithm != "sha256" {
		t.Fatalf("expected PasswordAlgorithm default sha256, got %s", cfg.Configuration.Auth.PasswordAlgorithm)
	}
}

func TestMapValueToConfigAuthSection(t *testing.T) {
	cfg := GetDefaultConfig()

	values := []database.ConfigValue{
		{Section: "Configuration.Auth", Key: "MaxFailedAttempts", Value: "7", ValueType: "int"},
		{Section: "Configuration.Auth", Key: "AccountLockMinutes", Value: "30", ValueType: "int"},
		{Section: "Configuration.Auth", Key: "PasswordAlgorithm", Value: "argon2id", ValueType: "string"},
	}

	for _, v := range values {
		mapValueToConfig(cfg, v)
	}

	if cfg.Configuration.Auth.MaxFailedAttempts != 7 {
		t.Fatalf("expected MaxFailedAttempts to be 7 after mapping, got %d", cfg.Configuration.Auth.MaxFailedAttempts)
	}
	if cfg.Configuration.Auth.AccountLockMinutes != 30 {
		t.Fatalf("expected AccountLockMinutes to be 30 after mapping, got %d", cfg.Configuration.Auth.AccountLockMinutes)
	}
	if cfg.Configuration.Auth.PasswordAlgorithm != "argon2id" {
		t.Fatalf("expected PasswordAlgorithm to be argon2id after mapping, got %s", cfg.Configuration.Auth.PasswordAlgorithm)
	}
}

func TestConfigToValuesIncludesAuth(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Configuration.Auth.MaxFailedAttempts = 8
	cfg.Configuration.Auth.AccountLockMinutes = 60
	cfg.Configuration.Auth.PasswordAlgorithm = "bcrypt"

	values := configToValues(cfg)

	assertValue := func(key string, expected string) {
		t.Helper()
		for _, v := range values {
			if v.Section == "Configuration.Auth" && v.Key == key {
				if v.Value != expected {
					t.Fatalf("expected %s for key %s, got %s", expected, key, v.Value)
				}
				return
			}
		}
		t.Fatalf("missing Configuration.Auth entry for key %s", key)
	}

	assertValue("MaxFailedAttempts", "8")
	assertValue("AccountLockMinutes", "60")
	assertValue("PasswordAlgorithm", "bcrypt")
}
