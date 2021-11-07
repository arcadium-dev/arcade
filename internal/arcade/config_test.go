package arcade

import (
	"testing"

	"arcadium.dev/core/config"
	"arcadium.dev/core/test"
)

func TestConfig(t *testing.T) {
	e := test.Env(map[string]string{
		// Log config
		"ARCADE_LOG_LEVEL":  "debug",
		"ARCADE_LOG_FORMAT": "JSON",

		// DB confing
		"ARCADE_POSTGRES_DB":   "arcade",
		"ARCADE_POSTGRES_HOST": "postgres-host",

		// Server config
		"ARCADE_SERVER_ADDR": ":4201",
	})
	e.Set(t)

	cfg, err := NewConfig(config.WithPrefix("arcade"))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if cfg == nil {
		t.Error("Expected a valid config")
	}

	t.Run("Test Log", func(t *testing.T) {
		l := cfg.Logger
		if l.Level() != "debug" {
			t.Errorf("Unexpected log level: %s", l.Level())
		}
		if l.Format() != "JSON" {
			t.Errorf("Unexpected log format: %s", l.Format())
		}
	})

	t.Run("Test DB", func(t *testing.T) {
		db := cfg.DB
		expectedDSN := "postgres://postgres-host/arcade?sslmode=verify-full"
		if db.DSN() != expectedDSN {
			t.Errorf("\nExpected DSN: %s\nActuual DSN:  %s", expectedDSN, db.DSN())
		}
	})

	t.Run("Test Server", func(t *testing.T) {
		server := cfg.Server
		if server.Addr() != ":4201" {
			t.Errorf("Unexpected server address: %s", server.Addr())
		}
		if server.Cert() != "" {
			t.Errorf("Unexpected server cert: %s", server.Cert())
		}
		if server.Key() != "" {
			t.Errorf("Unexpected server key: %s", server.Key())
		}
		if server.CACert() != "" {
			t.Errorf("Unexpected server cacert: %s", server.CACert())
		}
	})
}
