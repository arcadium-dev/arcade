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

package main

import (
	"context"
	"fmt"
	"io"
	l "log"
	"os/signal"
	"path/filepath"
	"runtime"

	"os"
	"sync"

	"arcadium.dev/core/build"
	"arcadium.dev/core/config"
	chttp "arcadium.dev/core/http"
	"arcadium.dev/core/log"
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade/http"
	"arcadium.dev/arcade/storage"
	"arcadium.dev/arcade/storage/cockroach"
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

		config Config
		logger log.Logger
		db     *sql.DB

		apiWG       sync.WaitGroup // To ensure stop isn't called before Start is ready.
		apiServices []chttp.Service
		apiServer   *chttp.Server

		telemetryWG       sync.WaitGroup // To ensure stop isn't called before Start is ready.
		telemetryServices []chttp.Service
		telemetryServer   *chttp.Server

		Stdout, Stderr io.Writer    // Provides a way for unit tests to capture output to standard file descriptors.
		Constructors   Constructors // Provides a way for unit tests to inject different object constructors.
	}

	// Constructors provide a way to inject different functions to create server components.
	Constructors struct {
		NewConfig          func(...config.Option) (Config, error)
		NewLogger          func(LoggerConfig) (log.Logger, error)
		NewDB              func(DBConfig, log.Logger) (*sql.DB, error)
		NewAPIServer       func(ServerConfig, TLSConfig, log.Logger, ...chttp.ServerOption) (*chttp.Server, error)
		NewTelemetryServer func(ServerConfig, TLSConfig, log.Logger, ...chttp.ServerOption) (*chttp.Server, error)
	}
)

// NewServer returns a new assets server.
func NewServer() *Server {
	Name = filepath.Base(os.Args[0])
	Go = runtime.Version()

	s := &Server{
		interrupt: make(chan os.Signal, 1),
		Constructors: Constructors{
			NewConfig: func(opts ...config.Option) (Config, error) {
				return NewConfig(opts...)
			},

			NewLogger: func(cfg LoggerConfig) (log.Logger, error) {
				return log.New(
					log.WithLevel(log.ToLevel(cfg.Level())),
					log.WithFormat(log.ToFormat(cfg.Format())),
					log.AsDefault(),
				)
			},

			NewDB: func(cfg DBConfig, logger log.Logger) (*sql.DB, error) {
				return sql.Open(cfg.Driver(), cfg.DSN(), logger)
			},

			NewAPIServer: func(cfg ServerConfig, tls TLSConfig, logger log.Logger, opts ...chttp.ServerOption) (*chttp.Server, error) {
				tlsConfig, err := tls.TLSConfig(config.WithMTLS())
				if err != nil {
					return nil, err
				}

				opts = append(opts,
					chttp.WithServerAddr(cfg.Addr()),
					chttp.WithServerTLS(tlsConfig),
					chttp.WithServerLogger(logger),
				)
				return chttp.NewServer(opts...), nil
			},

			NewTelemetryServer: func(cfg ServerConfig, tls TLSConfig, logger log.Logger, opts ...chttp.ServerOption) (*chttp.Server, error) {
				tlsConfig, err := tls.TLSConfig()
				if err != nil {
					return nil, err
				}

				opts = append(opts,
					chttp.WithServerAddr(cfg.Addr()),
					chttp.WithServerTLS(tlsConfig),
					chttp.WithServerLogger(logger),
				)
				return chttp.NewServer(opts...), nil
			},
		},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
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
	if len(args) == 1 && args[0] == "version" {
		fmt.Fprintln(s.Stdout, info)
		return
	}

	// Setup signal handler.
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(s.interrupt, os.Interrupt)
	go func() { <-s.interrupt; cancel() }()

	// Setup a temporary logger.
	lg := l.Default()
	lg.SetOutput(s.Stderr)

	// Load the config.
	s.config, err = s.Constructors.NewConfig()
	if err != nil {
		lg.Printf("error: failed to load config: %s", err)
		return
	}

	// Create a logger.
	s.logger, err = s.Constructors.NewLogger(s.config.Logger)
	if err != nil {
		lg.Printf("error: failed to create logger: %s", err)
		return
	}

	// Log the start.
	var start []interface{} = append([]interface{}{"msg", "starting"}, info.Fields()...)
	s.logger.Info(start...)

	// Setup database.
	s.db, err = s.Constructors.NewDB(s.config.DB, s.logger)
	if err != nil {
		s.logger.Error("msg", "failed to open db", "error", err)
		return
	}
	defer s.db.Close()

	// Setup API services.
	s.apiServices = []chttp.Service{
		http.PlayersService{Storage: storage.Players{DB: s.db.DB, Driver: cockroach.Driver{}}},
		http.RoomsService{Storage: storage.Rooms{DB: s.db.DB, Driver: cockroach.Driver{}}},
		http.LinksService{Storage: storage.Links{DB: s.db.DB, Driver: cockroach.Driver{}}},
		http.ItemsService{Storage: storage.Items{DB: s.db.DB, Driver: cockroach.Driver{}}},
	}

	// Setup telemetry services.
	s.telemetryServices = []chttp.Service{
		http.HealthService{},
		http.MetricsService{},
	}

	// Create ths API server.
	s.apiServer, err = s.Constructors.NewAPIServer(
		s.config.APIServer,
		s.config.TLS,
		s.logger,
		chttp.WithMiddleware(chttp.Metrics),
	)
	if err != nil {
		s.logger.Error("msg", "failed to create api server", "error", err)
		return
	}
	s.apiServer.Register(s.apiServices...)

	// Create the telemetry server.
	s.telemetryServer, err = s.Constructors.NewTelemetryServer(
		s.config.TelemetryServer,
		s.config.TLS,
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
