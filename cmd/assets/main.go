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
	"time"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"

	"arcadium.dev/core/http/middleware"
	httpserver "arcadium.dev/core/http/server"
	"arcadium.dev/core/http/services"
	"arcadium.dev/core/mpserver"
	"arcadium.dev/core/sql"

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
		return fmt.Errorf("failed to load mpserver configuration: %w", err)
	}
	httpCfg, err := httpserver.NewConfig("assets_http")
	if err != nil {
		return fmt.Errorf("failed to load http server configuration: %w", err)
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

	svcs := []httpserver.Service{
		services.Health{Start: time.Now(), Info: s.Info()},
		services.Metrics{},
	}
	// if httpCfg.PProfEnabled() {
	// svcs = append(svcs, services.PProf{})
	// }

	assetSvcs, err := createServices(ctx, cfg)
	if err != nil {
		return err
	}
	svcs = append(svcs, assetSvcs...)

	logger := zerolog.Ctx(ctx)
	mw := []mux.MiddlewareFunc{
		middleware.Recover{Logger: logger}.Panics,
		middleware.Logging{Logger: logger}.Requests,
		middleware.Metrics,
	}
	httpServer.Middleware(mw...)

	httpServer.Register(ctx, svcs...)

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

func createServices(ctx context.Context, cfg Config) ([]httpserver.Service, error) {
	switch cfg.Database {
	case postgresDatabase:
		return createPostgresServices(ctx, cfg)
	}

	return nil, fmt.Errorf("unknown database configured: %s", cfg.Database)
}

func createPostgresServices(ctx context.Context, cfg Config) ([]httpserver.Service, error) {
	var (
		db   *sql.DB
		err  error
		svcs []httpserver.Service
	)

	if cfg.PostgresDsn == "" {
		return nil, fmt.Errorf("postgres dsn required")
	}

	db, err = postgres.Open(ctx, cfg.PostgresDsn)
	if err != nil {
		return nil, err
	}

	items := server.ItemsService{
		Storage: data.ItemStorage{
			DB: db,
			Driver: postgres.ItemDriver{
				Driver: postgres.Driver{},
			},
		},
	}
	svcs = append(svcs, items)

	links := server.LinksService{
		Storage: data.LinkStorage{
			DB: db,
			Driver: postgres.LinkDriver{
				Driver: postgres.Driver{},
			},
		},
	}
	svcs = append(svcs, links)

	players := server.PlayersService{
		Storage: data.PlayerStorage{
			DB: db,
			Driver: postgres.PlayerDriver{
				Driver: postgres.Driver{},
			},
		},
	}
	svcs = append(svcs, players)

	rooms := server.RoomsService{
		Storage: data.RoomStorage{
			DB: db,
			Driver: postgres.RoomDriver{
				Driver: postgres.Driver{},
			},
		},
	}
	svcs = append(svcs, rooms)

	return svcs, nil
}
