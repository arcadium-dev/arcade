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

package main

import (
	"crypto/tls"

	"arcadium.dev/core/config"
)

type (
	// Config contains the configuration of the server.
	Config struct {
		Logger          LoggerConfig
		SQL             SQLConfig
		TLS             TLSConfig
		TelemetryServer ServerConfig
	}

	LoggerConfig interface {
		Level() string
		Format() string
	}

	SQLConfig interface {
		Driver() string
		URL() string
	}

	TLSConfig interface {
		Cert() string
		Key() string
		CACert() string
		TLSConfig(...config.TLSOption) (*tls.Config, error)
	}

	ServerConfig interface {
		Addr() string
	}
)

// NewConfig returns the configuration of the server.
func NewConfig(opts ...cconfig.Option) (Config, error) {
	var err error
	c := Config{}
	if c.Logger, err = cconfig.NewLogger(opts...); err != nil {
		return Config{}, err
	}
	if c.SQL, err = cconfig.NewSQL(opts...); err != nil {
		return Config{}, err
	}
	if c.TLS, err = cconfig.NewTLS(opts...); err != nil {
		return Config{}, err
	}
	telemertyOpts := append(opts, cconfig.WithPrefix("telemetry"))
	if c.TelemetryServer, err = cconfig.NewServer(telemertyOpts...); err != nil {
		return Config{}, err
	}
	return c, nil
}
