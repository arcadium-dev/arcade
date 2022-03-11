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
	"crypto/tls"

	cfg "arcadium.dev/core/config"
)

type (
	// config contains the configuration of the server.
	config struct {
		logger          loggerConfig
		sql             sqlConfig
		tls             tlsConfig
		apiServer       serverConfig
		telemetryServer serverConfig
	}

	loggerConfig interface {
		Level() string
		Format() string
	}

	sqlConfig interface {
		Driver() string
		URL() string
	}

	tlsConfig interface {
		Cert() string
		Key() string
		CACert() string
		TLSConfig(...cfg.TLSOption) (*tls.Config, error)
	}

	serverConfig interface {
		Addr() string
	}
)

// newConfig returns the configuration of the server.
func newConfig(opts ...cfg.Option) (config, error) {
	var err error
	c := config{}
	if c.logger, err = cfg.NewLogger(opts...); err != nil {
		return config{}, err
	}
	if c.sql, err = cfg.NewSQL(opts...); err != nil {
		return config{}, err
	}
	if c.tls, err = cfg.NewTLS(opts...); err != nil {
		return config{}, err
	}
	apiOpts := append(opts, cfg.WithPrefix("api"))
	if c.apiServer, err = cfg.NewServer(apiOpts...); err != nil {
		return config{}, err
	}
	telemertyOpts := append(opts, cfg.WithPrefix("telemetry"))
	if c.telemetryServer, err = cfg.NewServer(telemertyOpts...); err != nil {
		return config{}, err
	}
	return c, nil
}
