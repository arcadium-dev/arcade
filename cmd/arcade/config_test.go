//  Copyright 2022 arcadium.dev <info@arcadium.dev>
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
	"testing"
)

func TestConfig(t *testing.T) {
	// Log config
	t.Setenv("LOG_LEVEL", "Debug")
	t.Setenv("LOG_FORMAT", "JSON")

	// DB confing
	t.Setenv("SQL_URL", "cockroachdb://arcadium@cockroah:26257/assets?sslmode=verify-full")

	// TLS config
	t.Setenv("TLS_CERT", "/etc/certs/cert.pem")
	t.Setenv("TLS_KEY", "/etc/certs/key.pem")
	t.Setenv("TLS_CACERT", "/etc/certs/rootCA.pem")

	// Server config
	t.Setenv("API_SERVER_ADDR", ":4201")
	t.Setenv("TELEMETRY_SERVER_ADDR", ":4202")

	cfg, err := assets.NewConfig()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	t.Run("Test Log", func(t *testing.T) {
		l := cfg.Logger
		if l.Level() != "debug" {
			t.Error("Unexpected log level")
		}
		if l.Format() != "json" {
			t.Errorf("Unexpected log format")
		}
	})

	t.Run("Test DB", func(t *testing.T) {
		sql := cfg.SQL
		expectedURL := "cockroachdb://arcadium@cockroah:26257/assets?sslmode=verify-full"
		if sql.URL() != expectedURL {
			t.Errorf("\nExpected URL: %s\nActuual URL:  %s", expectedURL, sql.URL())
		}
	})

	t.Run("Test TLS", func(t *testing.T) {
		tls := cfg.TLS
		if tls.Cert() != "/etc/certs/cert.pem" {
			t.Errorf("Unexpected shared server cert: %s", tls.Cert())
		}
		if tls.Key() != "/etc/certs/key.pem" {
			t.Errorf("Unexpected shared server key: %s", tls.Key())
		}
		if tls.CACert() != "/etc/certs/rootCA.pem" {
			t.Errorf("Unexpected shared server cacert: %s", tls.CACert())
		}
	})

	t.Run("Test Server", func(t *testing.T) {
		apiServer := cfg.APIServer
		if apiServer.Addr() != ":4201" {
			t.Errorf("Unexpected server address: %s", apiServer.Addr())
		}
		telemetryServer := cfg.TelemetryServer
		if telemetryServer.Addr() != ":4202" {
			t.Errorf("Unexpected server address: %s", telemetryServer.Addr())
		}
	})
}
