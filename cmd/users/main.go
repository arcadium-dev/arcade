// Copyright 2024 arcadium.dev <info@arcadium.dev>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	l "log"

	"github.com/kelseyhightower/envconfig"

	httpserver "arcadium.dev/core/http/server"
	"arcadium.dev/core/rest"

	"arcadium.dev/arcade/data"
	"arcadium.dev/arcade/data/postgres"
	"arcadium.dev/arcade/user/server"
)

var (
	Version string
	Branch  string
	Commit  string
	Date    string
)

type (
	// RestServer defines the expected behavior of a rest server.
	RestServer interface {
		Init(...string) error
		Start(...httpserver.Service) error
		Ctx() context.Context
	}

	Config struct {
		Database    string `required:"true"`
		PostgresDsn string `split_words:"true"`
	}
)

const (
	postgresDatabase = "postgres"
)

// New creates a new rest server. This is provided as a function variable to
// allow for easier unit testing.
var New = func(v, b, c, d string) RestServer {
	return rest.NewServer(v, b, c, d)
}

// Main is the testable entry point into the users server.
func Main() error {
	s := New(Version, Branch, Commit, Date)

	prefix := "users"

	cfg := Config{}
	if err := envconfig.Process(prefix, &cfg); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := s.Init(prefix); err != nil {
		return err
	}

	switch cfg.Database {
	case postgresDatabase:
		if err := startPostgres(s, cfg.PostgresDsn); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown database configured: %s", cfg.Database)
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		l.Fatal(err)
	}
}

func startPostgres(s RestServer, dsn string) error {
	if dsn == "" {
		return fmt.Errorf("postgres dsn required")
	}

	db, err := postgres.Open(s.Ctx(), dsn)
	if err != nil {
		return err
	}

	users := server.UsersService{
		Storage: data.UserStorage{
			DB: db,
			Driver: postgres.UserDriver{
				Driver: postgres.Driver{},
			},
		},
	}

	return s.Start(users)
}
