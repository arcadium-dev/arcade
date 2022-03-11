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

package arcade

import (
	"testing"

	"arcadium.dev/core/test"
)

func TestConfig(t *testing.T) {
	e := test.Env(map[string]string{
		// Log config
		"LOG_LEVEL":  "Debug",
		"LOG_FORMAT": "JSON",

		// TLS config
		"TLS_CERT":   "/etc/certs/cert.pem",
		"TLS_KEY":    "/etc/certs/key.pem",
		"TLS_CACERT": "/etc/certs/rootCA.pem",

		// Server config
		"TELEMETRY_SERVER_ADDR": ":4202",
	})
	e.Set(t)

	cfg, err := newConfig()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	t.Run("Test Log", func(t *testing.T) {
		l := cfg.logger
		if l.Level() != "debug" {
			t.Error("Unexpected log level")
		}
		if l.Format() != "json" {
			t.Errorf("Unexpected log format")
		}
	})

	t.Run("Test TLS", func(t *testing.T) {
		tls := cfg.tls
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
		telemetryServer := cfg.telemetryServer
		if telemetryServer.Addr() != ":4202" {
			t.Errorf("Unexpected server address: %s", telemetryServer.Addr())
		}
	})
}
