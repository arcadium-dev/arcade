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

package assets

import (
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
	"arcadium.dev/core/test"
)

func TestServer(t *testing.T) {
	args := []string{"assets"}

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
		s.ctors.newConfig = func(...cconfig.Option) (config, error) {
			return config{}, errors.New("config construction failure")
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
		s.ctors.newConfig = func(...cconfig.Option) (config, error) {
			return config{
				logger: cconfig.Logger{},
			}, nil
		}
		s.ctors.newLogger = func(loggerConfig) (log.Logger, error) {
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
		s.ctors.newConfig = func(...cconfig.Option) (config, error) {
			return config{
				logger: mockLoggerConfig{level: "debug", format: "logfmt"},
				sql:    mockSQLConfig{driver: "pgx", url: "pgx://cockroach:26257/assets"},
			}, nil
		}

		s.ctors.newLogger = func(cfg loggerConfig) (log.Logger, error) {
			return log.New(
				log.WithLevel(log.ToLevel(cfg.Level())),
				log.WithFormat(log.ToFormat(cfg.Format())),
				log.WithOutput(b),
				log.WithoutTimestamp(),
			)
		}
		s.ctors.newDB = func(cfg sqlConfig, logger log.Logger) (*sql.DB, error) {
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
		s.ctors.newConfig = func(...cconfig.Option) (config, error) {
			return config{
				logger: mockLoggerConfig{level: "debug", format: "logfmt"},
			}, nil
		}

		s.ctors.newLogger = func(cfg loggerConfig) (log.Logger, error) {
			return log.New(
				log.WithLevel(log.ToLevel(cfg.Level())),
				log.WithFormat(log.ToFormat(cfg.Format())),
				log.WithOutput(b),
				log.WithoutTimestamp(),
			)
		}

		var m sqlmock.Sqlmock
		s.ctors.newDB = func(cfg sqlConfig, logger log.Logger) (*sql.DB, error) {
			db, mock, err := sqlmock.New()
			if db == nil || mock == nil || err != nil {
				t.Fatal("Failed to create sqlmock")
			}
			m = mock
			m.ExpectClose()
			return &sql.DB{DB: db}, err
		}

		// Create mock services here.

		s.ctors.newAPIServer = func(serverConfig, tlsConfig, log.Logger, ...http.ServerOption) (*http.Server, error) {
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
		// TODO
	})

	t.Run("success", func(t *testing.T) {
		// TODO
	})
}

func setup() (*Server, *test.StringBuffer) {
	s := New("assets", "version", "branch", "commit", "date", "go")
	b := test.NewStringBuffer()
	s.stdout = b
	s.stderr = b

	return s, b
}

type (
	mockLoggerConfig struct {
		level, format string
	}

	mockSQLConfig struct {
		driver, url string
	}
)

func (m mockLoggerConfig) Level() string {
	return m.level
}

func (m mockLoggerConfig) Format() string {
	return m.format
}

func (m mockSQLConfig) Driver() string {
	return m.driver
}

func (m mockSQLConfig) URL() string {
	return m.url
}
