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
	"context"
	"fmt"
	"io"
	l "log"
	"os/signal"

	"os"
	"sync"

	"arcadium.dev/core/build"
	cconfig "arcadium.dev/core/config"
	"arcadium.dev/core/http"
	"arcadium.dev/core/log"
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade/internal/health"
	"arcadium.dev/arcade/internal/links"
	"arcadium.dev/arcade/internal/metrics"
	"arcadium.dev/arcade/internal/players"
	"arcadium.dev/arcade/internal/rooms"
)

// Build information.
var (
	Name    string
	Version string
	Branch  string
	Commit  string
	Date    string
	Go      string
)

type (
	// Server represents the assets server.
	Server struct {
		interrupt chan os.Signal

		config config
		logger log.Logger
		db     *sql.DB

		apiWG       sync.WaitGroup // To ensure stop isn't called before Start is ready.
		apiServices []http.Service
		apiServer   *http.Server

		telemetryWG       sync.WaitGroup // To ensure stop isn't called before Start is ready.
		telemetryServices []http.Service
		telemetryServer   *http.Server

		ctors          constructors // Provides a way for unit tests to inject different object constructors.
		stdout, stderr io.Writer    // Provides a way for unit tests to capture output to standard file descriptors.
	}

	// constructors provide a way to inject different functions to create server components.
	constructors struct {
		newConfig          func(...cconfig.Option) (config, error)
		newLogger          func(loggerConfig) (log.Logger, error)
		newDB              func(sqlConfig, log.Logger) (*sql.DB, error)
		newAPIServer       func(serverConfig, tlsConfig, log.Logger, ...http.ServerOption) (*http.Server, error)
		newTelemetryServer func(serverConfig, tlsConfig, log.Logger, ...http.ServerOption) (*http.Server, error)
	}
)

// New returns a new assets server.
func New(name, version, branch, commit, date, gover string) *Server {
	Name = name
	Version = version
	Branch = branch
	Commit = commit
	Date = date
	Go = gover

	s := &Server{
		interrupt: make(chan os.Signal, 1),
		ctors: constructors{
			newConfig: func(opts ...cconfig.Option) (config, error) {
				return newConfig(opts...)
			},

			newLogger: func(cfg loggerConfig) (log.Logger, error) {
				return log.New(
					log.WithLevel(log.ToLevel(cfg.Level())),
					log.WithFormat(log.ToFormat(cfg.Format())),
					log.AsDefault(),
				)
			},

			newDB: func(cfg sqlConfig, logger log.Logger) (*sql.DB, error) {
				return sql.Open(cfg.Driver(), cfg.URL(), logger)
			},

			newAPIServer: func(cfg serverConfig, tls tlsConfig, logger log.Logger, opts ...http.ServerOption) (*http.Server, error) {
				tlsConfig, err := tls.TLSConfig(cconfig.WithMTLS())
				if err != nil {
					return nil, err
				}

				opts = append(opts,
					http.WithServerAddr(cfg.Addr()),
					http.WithServerTLS(tlsConfig),
					http.WithServerLogger(logger),
				)
				return http.NewServer(opts...), nil
			},

			newTelemetryServer: func(cfg serverConfig, tls tlsConfig, logger log.Logger, opts ...http.ServerOption) (*http.Server, error) {
				tlsConfig, err := tls.TLSConfig()
				if err != nil {
					return nil, err
				}

				opts = append(opts,
					http.WithServerAddr(cfg.Addr()),
					http.WithServerTLS(tlsConfig),
					http.WithServerLogger(logger),
				)
				return http.NewServer(opts...), nil
			},
		},
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	s.apiWG.Add(1)
	s.telemetryWG.Add(1)

	return s
}

// Start is the entry point into the service.
func (s *Server) Start(args []string) {
	var err error

	defer func() {
		if err != nil {
			s.apiWG.Done()
			s.telemetryWG.Done()
		}
	}()

	info := build.Info(Name, Version, Branch, Commit, Date)

	// Return the version when given a "version" argument.
	if len(args) == 2 && args[1] == "version" {
		fmt.Fprintln(s.stdout, info)
		return
	}

	// Setup signal handler.
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(s.interrupt, os.Interrupt)
	go func() { <-s.interrupt; cancel() }()

	// Setup a temporary logger.
	lg := l.Default()
	lg.SetOutput(s.stderr)

	// Load the config.
	s.config, err = s.ctors.newConfig()
	if err != nil {
		lg.Printf("error: failed to load config: %s", err)
		return
	}

	// Create a logger.
	s.logger, err = s.ctors.newLogger(s.config.logger)
	if err != nil {
		lg.Printf("error: failed to create logger: %s", err)
		return
	}

	// Log the start.
	var start []interface{} = append([]interface{}{"msg", "starting"}, info.Fields()...)
	s.logger.Info(start...)

	// Setup database.
	s.db, err = s.ctors.newDB(s.config.sql, s.logger)
	if err != nil {
		s.logger.Error("msg", "failed to open db", "error", err)
		return
	}
	defer s.db.Close()

	// Setup API services.
	s.apiServices = []http.Service{
		players.New(s.db.DB),
		rooms.New(s.db.DB),
		links.New(s.db.DB),
		// items.New(s.db.DB)
	}

	// Setup telemetry services.
	s.telemetryServices = []http.Service{
		health.Service{},
		metrics.Service{},
	}

	// Create ths API server.
	s.apiServer, err = s.ctors.newAPIServer(
		s.config.apiServer,
		s.config.tls,
		s.logger,
		http.WithMiddleware(http.Metrics),
	)
	if err != nil {
		s.logger.Error("msg", "failed to create api server", "error", err)
		return
	}
	s.apiServer.Register(s.apiServices...)

	// Create the telemetry server.
	s.telemetryServer, err = s.ctors.newTelemetryServer(
		s.config.telemetryServer,
		s.config.tls,
		s.logger,
	)
	if err != nil {
		s.logger.Error("msg", "failed to create telemetry server", "error", err)
		return
	}
	s.telemetryServer.Register(s.telemetryServices...)

	// Serve.
	apiResult := make(chan error, 1)
	go func() {
		s.apiWG.Done()
		apiResult <- s.apiServer.Serve()
	}()

	telemetryResult := make(chan error, 1)
	go func() {
		s.telemetryWG.Done()
		telemetryResult <- s.telemetryServer.Serve()
	}()

	select {
	// Wait for an interrupt.
	case <-ctx.Done():
		s.apiShutdown()
		s.telemetryShutdown()

	// If the apiServer failed to Serve, log the error and return failure..
	case err = <-apiResult:
		s.telemetryShutdown()
		s.logger.Error("msg", "failed to start api server", "error", err)

	// If the telemetryServer failed to Serve, log the error and return failure..
	case err = <-telemetryResult:
		s.apiShutdown()
		s.logger.Error("msg", "failed to start telemetry server", "error", err)
	}
}

// Stop halts the server.
func (s *Server) Stop() {
	s.apiWG.Wait()
	s.telemetryWG.Wait()
	close(s.interrupt)
}

func (s *Server) apiShutdown() {
	for _, service := range s.apiServices {
		service.Shutdown()
	}
	s.apiServer.Shutdown()
}

func (s *Server) telemetryShutdown() {
	for _, service := range s.telemetryServices {
		service.Shutdown()
	}
	s.telemetryServer.Shutdown()
}
