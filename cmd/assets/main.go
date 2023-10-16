// Copyright 2021-2023 arcadium.dev <info@arcadium.dev>
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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kelseyhightower/envconfig"

	httpserver "arcadium.dev/core/http/server"
	"arcadium.dev/core/rest"

	"arcadium.dev/arcade/assets/data"
	"arcadium.dev/arcade/assets/data/cockroach"
	"arcadium.dev/arcade/assets/rest/server"
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
		DSN string `required:"true"`
	}
)

// New creates a new rest server. This is provided as a function variable to
// allow for easier unit testing.
var New = func(v, b, c, d string) RestServer {
	return rest.NewServer(v, b, c, d)
}

// Main is the testable entry point into the assets server.
func Main() error {
	s := New(Version, Branch, Commit, Date)

	prefix := "assets"

	cfg := Config{}
	if err := envconfig.Process(prefix, &cfg); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := s.Init(prefix); err != nil {
		return err
	}

	db, err := cockroach.Open(s.Ctx(), cfg.DSN)
	if err != nil {
		return err
	}

	items := server.ItemsService{
		Storage: data.ItemStorage{
			DB: db,
			Driver: cockroach.ItemDriver{
				Driver: cockroach.Driver{},
			},
		},
	}

	links := server.LinksService{
		Storage: data.LinkStorage{
			DB: db,
			Driver: cockroach.LinkDriver{
				Driver: cockroach.Driver{},
			},
		},
	}

	players := server.PlayersService{
		Storage: data.PlayerStorage{
			DB: db,
			Driver: cockroach.PlayerDriver{
				Driver: cockroach.Driver{},
			},
		},
	}

	rooms := server.RoomsService{
		Storage: data.RoomStorage{
			DB: db,
			Driver: cockroach.RoomDriver{
				Driver: cockroach.Driver{},
			},
		},
	}

	if err := s.Start(items, links, players, rooms); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		l.Fatal(err)
	}
}
