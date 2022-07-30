//  Copyright 2021-2022 arcadium.dev <info@arcadium.dev>
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main_test

import (
	"crypto/tls"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	cconfig "arcadium.dev/core/config"
	"arcadium.dev/core/http"
	"arcadium.dev/core/log"
	"arcadium.dev/core/sql"

	assets "arcadium.dev/arcade/cmd/assets"
)

func TestServer(t *testing.T) {
	args := []string{}

	t.Run("version", func(t *testing.T) {
		s, b := setup()
		vargs := append(args, "version")

		s.Start(vargs)
		if b.Len() != 1 {
			t.Errorf("Unexpected version buffer length: %d", b.Len())
		}
		expected := fmt.Sprintf("assets version (branch: branch, commit: commit, date: date, go: %s)\n", runtime.Version())
		if b.Index(0) != expected {
			t.Errorf("\nExpected version: %s\nActual version:   %s", expected, b.Index(0))
		}
	})

	t.Run("config construction failure", func(t *testing.T) {
		s, b := setup()
		s.Constructors.NewConfig = func(...cconfig.Option) (assets.Config, error) {
			return assets.Config{}, errors.New("config construction failure")
		}

		s.Start(args)
		if b.Len() != 1 {
			t.Fatalf("Unexpected error log buffer length: %d", b.Len())
		}
		expected := "failed to load config: config construction failure\n"
		if !strings.Contains(b.Index(0), expected) {
			t.Errorf("\nExpected error log: %s\nActual error log:   %s", expected, b.Index(0))
		}
	})

	t.Run("log construction failure", func(t *testing.T) {
		s, b := setup()
		s.Constructors.NewConfig = func(...cconfig.Option) (assets.Config, error) {
			return assets.Config{
				Logger: cconfig.Logger{},
			}, nil
		}
		s.Constructors.NewLogger = func(assets.LoggerConfig) (log.Logger, error) {
			return log.Logger{}, errors.New("log construction failure")
		}

		s.Start(args)
		if b.Len() != 1 {
			t.Errorf("Unexpected error log buffer length: %d", b.Len())
		}
		expected := "failed to create logger: log construction failure\n"
		if !strings.Contains(b.Index(0), expected) {
			t.Errorf("\nExpected error log: %s\nActual error log:   %s", expected, b.Index(0))
		}
	})

	t.Run("db construction failure", func(t *testing.T) {
		s, b := setup()
		s.Constructors.NewConfig = func(...cconfig.Option) (assets.Config, error) {
			return assets.Config{
				Logger: mockLoggerConfig{level: "debug", format: "logfmt"},
				DB:     mockDBConfig{driver: "pgx", dsn: "pgx://cockroach:26257/assets"},
			}, nil
		}

		s.Constructors.NewLogger = func(cfg assets.LoggerConfig) (log.Logger, error) {
			return log.New(
				log.WithLevel(log.ToLevel(cfg.Level())),
				log.WithFormat(log.ToFormat(cfg.Format())),
				log.WithOutput(b),
				log.WithoutTimestamp(),
			)
		}
		s.Constructors.NewDB = func(cfg assets.DBConfig, logger log.Logger) (*sql.DB, error) {
			return nil, errors.New("db construction failure")
		}

		s.Start(args)
		if b.Len() != 2 {
			t.Fatalf("Unexpected error log buffer length: %d", b.Len())
		}
		expected := `level=error msg="failed to open db" error="db construction failure"`
		if !strings.Contains(b.Index(1), expected) {
			t.Errorf("\nExpected error log: %s\nActual error log:   %s", expected, b.Index(1))
		}
	})

	t.Run("api server construction failure", func(t *testing.T) {
		s, b := setup()
		s.Constructors.NewConfig = func(...cconfig.Option) (assets.Config, error) {
			return assets.Config{
				Logger: mockLoggerConfig{level: "debug", format: "logfmt"},
			}, nil
		}

		s.Constructors.NewLogger = func(cfg assets.LoggerConfig) (log.Logger, error) {
			return log.New(
				log.WithLevel(log.ToLevel(cfg.Level())),
				log.WithFormat(log.ToFormat(cfg.Format())),
				log.WithOutput(b),
				log.WithoutTimestamp(),
			)
		}

		var m sqlmock.Sqlmock
		s.Constructors.NewDB = func(cfg assets.DBConfig, logger log.Logger) (*sql.DB, error) {
			db, mock, err := sqlmock.New()
			if db == nil || mock == nil || err != nil {
				t.Fatal("Failed to create sqlmock")
			}
			m = mock
			m.ExpectClose()
			return &sql.DB{DB: db}, err
		}

		s.Constructors.NewAPIServer = func(assets.ServerConfig, assets.TLSConfig, log.Logger, ...http.ServerOption) (*http.Server, error) {
			return nil, errors.New("api server construction failure")
		}

		s.Start(args)
		if b.Len() != 2 {
			t.Errorf("Unexpected error log buffer length: %d", b.Len())
		}
		expected := `level=error msg="failed to create api server" error="api server construction failure"`
		if !strings.Contains(b.Index(1), expected) {
			t.Errorf("\nExpected error log: %s\nActual error log:   %s", expected, b.Index(1))
		}

		if err := m.ExpectationsWereMet(); err != nil {
			t.Errorf("Failed to close sqlmock: %s", err)
		}
	})

	t.Run("telemetry server construction failure", func(t *testing.T) {
		s, b := setup()

		s.Constructors.NewConfig = func(...cconfig.Option) (assets.Config, error) {
			return assets.Config{
				Logger:    mockLoggerConfig{level: "debug", format: "logfmt"},
				TLS:       mockTLSConfig{},
				APIServer: mockServerConfig{addr: ":4201"},
			}, nil
		}

		s.Constructors.NewLogger = func(cfg assets.LoggerConfig) (log.Logger, error) {
			return log.New(
				log.WithLevel(log.ToLevel(cfg.Level())),
				log.WithFormat(log.ToFormat(cfg.Format())),
				log.WithOutput(b),
				log.WithoutTimestamp(),
			)
		}

		var m sqlmock.Sqlmock
		s.Constructors.NewDB = func(assets.DBConfig, log.Logger) (*sql.DB, error) {
			db, mock, err := sqlmock.New()
			if db == nil || mock == nil || err != nil {
				t.Fatal("Failed to create sqlmock")
			}
			m = mock
			m.ExpectClose()
			return &sql.DB{DB: db}, err
		}

		s.Constructors.NewTelemetryServer = func(assets.ServerConfig, assets.TLSConfig, log.Logger, ...http.ServerOption) (*http.Server, error) {
			return nil, errors.New("telemetry server construction failure")
		}

		s.Start(args)
		if b.Len() != 7 {
			t.Fatalf("Unexpected error log buffer length: %d", b.Len())
		}
		expected := `level=error msg="failed to create telemetry server" error="telemetry server construction failure"`
		if !strings.Contains(b.Index(6), expected) {
			t.Errorf("\nExpected error log: %s\nActual error log:   %s", expected, b.Index(6))
		}

		if err := m.ExpectationsWereMet(); err != nil {
			t.Errorf("Failed to close sqlmock: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		s := assets.NewServer()
		s.Constructors.NewDB = func(assets.DBConfig, log.Logger) (*sql.DB, error) {
			db, _, err := sqlmock.New()
			return &sql.DB{DB: db}, err
		}
		t.Setenv("API_SERVER_ADDR", ":4201")
		t.Setenv("TELEMETRY_SERVER_ADDR", ":4202")

		t.Setenv("TLS_CERT", "./insecure/assets.pem")
		t.Setenv("TLS_KEY", "./insecure/assets_key.pem")
		t.Setenv("TLS_CACERT", "./insecure/rootCA.pem")

		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("LOG_FORMAT", "logfmt")

		t.Setenv("DB_DSN", "postgresql://arcadium@cockroach:26257/arcade?sslmode-verify-full&sslrootcert=%2Fetc%2Fcerts%2Fca.crt&sslcert=%2Fetc%2Fcerts%2Fclient.arcadium.crt&sslkey=%2Fetc%2Fcerts%2Fclient.arcadium.key")

		r := make(chan struct{}, 1)
		go func() { s.Start(args); close(r) }()
		s.Stop()
		<-r
	})
}

func setup() (*assets.Server, *log.StringBuffer) {
	s := assets.NewServer()
	assets.Name = "assets"
	assets.Version = "version"
	assets.Branch = "branch"
	assets.Commit = "commit"
	assets.Date = "date"

	b := log.NewStringBuffer()
	s.Stdout = b
	s.Stderr = b

	return s, b
}

type (
	mockLoggerConfig struct {
		level, format string
	}

	mockDBConfig struct {
		driver, dsn string
	}

	mockServerConfig struct {
		addr string
	}

	mockTLSConfig struct {
		cert, key, cacert string
	}
)

func (m mockLoggerConfig) Level() string  { return m.level }
func (m mockLoggerConfig) Format() string { return m.format }

func (m mockDBConfig) Driver() string { return m.driver }
func (m mockDBConfig) DSN() string    { return m.dsn }

func (m mockServerConfig) Addr() string { return m.addr }

func (m mockTLSConfig) Cert() string   { return m.cert }
func (m mockTLSConfig) Key() string    { return m.key }
func (m mockTLSConfig) CACert() string { return m.cacert }
func (m mockTLSConfig) TLSConfig(...cconfig.TLSOption) (*tls.Config, error) {
	return &tls.Config{}, nil
}
