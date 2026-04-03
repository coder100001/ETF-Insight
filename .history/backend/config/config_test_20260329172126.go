package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if cfg.Server.Host == "" {
		t.Error("Server host should not be empty")
	}

	if cfg.Server.Port == 0 {
		t.Error("Server port should not be zero")
	}

	if cfg.Database.DSN == "" {
		t.Error("Database DSN should not be empty")
	}

	if len(cfg.ETF.DefaultSymbols) == 0 {
		t.Error("ETF default symbols should not be empty")
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	cfg, err := LoadConfig("nonexistent.yaml")
	if err != nil {
		t.Errorf("LoadConfig should not return error for non-existent file: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig returned nil config")
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Errorf("LoadConfig should not return error for empty path: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig returned nil config")
	}
}

func TestGetEnv(t *testing.T) {
	key := "TEST_ENV_VAR"
	defaultValue := "default"

	value := getEnv(key, defaultValue)
	if value != defaultValue {
		t.Errorf("Expected %s, got %s", defaultValue, value)
	}

	os.Setenv(key, "custom")
	defer os.Unsetenv(key)

	value = getEnv(key, defaultValue)
	if value != "custom" {
		t.Errorf("Expected custom, got %s", value)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	key := "TEST_INT_VAR"
	defaultValue := 8080

	value := getEnvAsInt(key, defaultValue)
	if value != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, value)
	}

	os.Setenv(key, "9090")
	defer os.Unsetenv(key)

	value = getEnvAsInt(key, defaultValue)
	if value != 9090 {
		t.Errorf("Expected 9090, got %d", value)
	}

	os.Setenv(key, "invalid")
	defer os.Unsetenv(key)

	value = getEnvAsInt(key, defaultValue)
	if value != defaultValue {
		t.Errorf("Expected %d for invalid value, got %d", defaultValue, value)
	}
}

func TestServerConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.ReadTimeout == 0 {
		t.Error("ReadTimeout should not be zero")
	}

	if cfg.Server.WriteTimeout == 0 {
		t.Error("WriteTimeout should not be zero")
	}
}

func TestETFConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ETF.DataFetch.RetryTimes == 0 {
		t.Error("RetryTimes should not be zero")
	}

	if cfg.ETF.DataFetch.RequestTimeout == 0 {
		t.Error("RequestTimeout should not be zero")
	}

	if cfg.ETF.Cache.RealtimeTTL == 0 {
		t.Error("RealtimeTTL should not be zero")
	}
}

func TestScheduleConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Schedule.DailyUpdateTime == "" {
		t.Error("DailyUpdateTime should not be empty")
	}

	if cfg.Schedule.Timezone == "" {
		t.Error("Timezone should not be empty")
	}
}
