// Copyright 2026 arcadium.dev <info@arcadium.dev>
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
	"fmt"
	l "log"
	"time"

	"github.com/gorilla/mux"

	"arcadium.dev/core/http/middleware"
	httpserver "arcadium.dev/core/http/server"
	"arcadium.dev/core/http/services"
	"arcadium.dev/core/mpserver"
	"arcadium.dev/core/telnet"

	"arcadium.dev/arcade/tad"
)

var (
	Version string
	Branch  string
	Commit  string
	Date    string
)

func Main() error {
	var err error

	// Create the multiprotocol server.
	mpCfg, err := mpserver.NewConfig("tad")
	if err != nil {
		return fmt.Errorf("failed to load mpserver configuration: %w", err)
	}

	s, err := mpserver.New(Version, Branch, Commit, Date, mpCfg.ToOptions()...)
	if err != nil {
		return fmt.Errorf("failed to create new mpserver: %w", err)
	}
	logger := s.Logger()

	// Create the telnet adapter server.
	telnetCfg, err := telnet.NewServerConfig("tad_telnet")
	if err != nil {
		return fmt.Errorf("failed to load http server configuration: %w", err)
	}

	telnetServer := telnet.NewServer(append(telnetCfg.ToOptions(), telnet.WithServerLogger(logger))...)
	telnetServer.Register(&tad.Service{})
	s.Register(telnetServer)

	// Create the http server, for the health and metrics services.
	httpCfg, err := httpserver.NewConfig("tad_http")
	if err != nil {
		return fmt.Errorf("failed to load http server configuration: %w", err)
	}

	httpServer, err := httpserver.New(append(httpCfg.ToOptions(), httpserver.WithLogger(logger))...)
	if err != nil {
		return fmt.Errorf("failed to create new http server: %w", err)
	}

	svcs := []httpserver.Service{
		services.Health{Start: time.Now(), Info: s.Info()},
		services.Metrics{},
	}

	mw := []mux.MiddlewareFunc{
		middleware.Recover{Logger: logger}.Panics,
		middleware.Logging{Logger: logger}.Requests,
		middleware.Metrics,
	}
	httpServer.Middleware(mw...)

	httpServer.Register(svcs...)
	s.Register(httpServer)

	// Let's go...
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
