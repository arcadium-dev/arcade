package arcade

import (
	"context"
	"fmt"
	"io"
	l "log"
	"os"
	"os/signal"
	"sync"

	"arcadium.dev/core/build"
	"arcadium.dev/core/config"
	"arcadium.dev/core/log"
)

// Return codes.
const (
	success int = 0
	failure int = 1
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
	// Server represents the arcade server.
	Server struct {
		interrupt chan os.Signal

		config *Config
		logger config.Logger

		ctors          constructors // Provides a way for unit tests to inject different objcet constructors.
		stdout, stderr io.Writer    // Provides a way for unit tests to capture output to standard file descriptors.

		wg sync.WaitGroup // To ensure Stop isn't called before Start is ready.
	}

	// constructors provide a way to inject different functions to create server components.
	constructors struct {
		NewConfig func(...config.Option) (*Config, error)
		NewLogger func(config.Logger) (log.Logger, error)
	}
)

// New returns a new arcade server.
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
			NewConfig: func(opts ...config.Option) (*Config, error) {
				return NewConfig(opts...)
			},
			NewLogger: func(cfg config.Logger) (log.Logger, error) {
				return log.New(
					log.WithLevel(log.ToLevel(cfg.Level())),
					log.WithFormat(log.ToFormat(cfg.Format())),
				)
			},
		},
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	s.wg.Add(1)

	return s
}

// Start is the entry point into the service.
func (s *Server) Start(args []string) int {
	var err error

	info := build.Info(Name, Version, Branch, Commit, Date)

	// Return the version when given a "version" argument.
	if len(args) == 1 && args[0] == "version" {
		fmt.Fprintln(s.stdout, info)
		return success
	}

	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(s.interrupt, os.Interrupt)
	go func() { <-s.interrupt; cancel() }()

	// Setup a temporary logger.
	lg := l.Default()
	lg.SetOutput(s.stderr)

	// Load the config.
	s.config, err = s.ctors.NewConfig(config.WithPrefix(Name))
	if err != nil {
		lg.Printf("Error: Failed to load config: %s", err)
		return failure
	}

	// Create a logger.
	logger, err := s.ctors.NewLogger(s.config.Logger)
	if err != nil {
		lg.Printf("Error: Failed to create logger: %s", err)
		return failure
	}

	// Log the start.
	var start []interface{} = append([]interface{}{"msg", "starting"}, info.Fields()...)
	logger.Info(start...)

	// Setup services.

	// Create server.

	// Wait for an interrupt.
	s.wg.Done()
	<-ctx.Done()

	return success
}

// Stop halts the server.
func (s *Server) Stop() {
	s.wg.Wait()
	close(s.interrupt)
}
