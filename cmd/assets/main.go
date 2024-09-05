// Copyright 2021-2024 arcadium.dev <info@arcadium.dev>
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

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"

	"arcadium.dev/core/http/middleware"
	httpserver "arcadium.dev/core/http/server"
	"arcadium.dev/core/mpserver"

	"arcadium.dev/arcade/asset/rest/server"
	"arcadium.dev/arcade/data"
	"arcadium.dev/arcade/data/postgres"
)

var (
	Version string
	Branch  string
	Commit  string
	Date    string
)

type (
	// Server defines the expected behavior of a server.
	Server interface {
		Serve() error
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

// Main is the testable entry point into the assets server.
func Main() error {
	cfg := Config{}
	if err := envconfig.Process("assets", &cfg); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	mpCfg, err := mpserver.NewConfig("assets")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	httpCfg, err := httpserver.NewConfig("assets_http")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s, err := mpserver.New(Version, Branch, Commit, Date, mpCfg.ToOptions()...)
	if err != nil {
		return fmt.Errorf("failed to create new mpserver: %w", err)
	}
	ctx := s.Ctx()

	httpServer, err := httpserver.New(ctx, httpCfg.ToOptions()...)
	if err != nil {
		return fmt.Errorf("failed to create new http server: %w", err)
	}

	s.Register(ctx, httpServer)

	switch cfg.Database {
	case postgresDatabase:
		if err := createPostgresServices(ctx, httpServer, cfg.PostgresDsn); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown database configured: %s", cfg.Database)
	}

	if err := s.Serve(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		l.Fatal(err)
	}
}

func createPostgresServices(ctx context.Context, s *httpserver.Server, dsn string) error {
	if dsn == "" {
		return fmt.Errorf("postgres dsn required")
	}

	db, err := postgres.Open(ctx, dsn)
	if err != nil {
		return err
	}

	items := server.ItemsService{
		Storage: data.ItemStorage{
			DB: db,
			Driver: postgres.ItemDriver{
				Driver: postgres.Driver{},
			},
		},
	}

	links := server.LinksService{
		Storage: data.LinkStorage{
			DB: db,
			Driver: postgres.LinkDriver{
				Driver: postgres.Driver{},
			},
		},
	}

	players := server.PlayersService{
		Storage: data.PlayerStorage{
			DB: db,
			Driver: postgres.PlayerDriver{
				Driver: postgres.Driver{},
			},
		},
	}

	rooms := server.RoomsService{
		Storage: data.RoomStorage{
			DB: db,
			Driver: postgres.RoomDriver{
				Driver: postgres.Driver{},
			},
		},
	}

	logger := zerolog.Ctx(ctx)
	mw := []mux.MiddlewareFunc{
		middleware.Recover{Logger: logger}.Panics,
		middleware.Logging{Logger: logger}.Requests,
		middleware.Metrics,
	}
	s.Middleware(mw...)

	s.Register(ctx, items, links, players, rooms)
	return nil
}
