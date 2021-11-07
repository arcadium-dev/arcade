package arcade

import (
	"fmt"
	"runtime"
	"strings"

	"testing"

	"arcadium.dev/core/config"
	"arcadium.dev/core/errors"
	"arcadium.dev/core/log"
	"arcadium.dev/core/test"

	"arcadium.dev/arcade/internal/arcade/mock"
)

func TestMain(t *testing.T) {
	args := []string{}

	t.Run("version", func(t *testing.T) {
		s := New("arcade", "version", "branch", "commit", "date", "go")
		vargs := append(args, "version")
		b := test.NewStringBuffer()
		s.stdout = b

		result := s.Start(vargs)
		if result != success {
			t.Errorf("Unexpected result: %d", result)
		}
		if len(b.Buffer) != 1 {
			t.Errorf("Unexpected version buffer length: %d", len(b.Buffer))
		}
		expected := fmt.Sprintf("arcade version (branch: branch, commit: commit, date: date, go: %s)\n", runtime.Version())
		if b.Buffer[0] != expected {
			t.Errorf("\nExpected version: %s\nActual version:   %s", expected, b.Buffer[0])
		}
	})

	t.Run("config construction failure", func(t *testing.T) {
		s := New("arcade", "version", "branch", "commit", "date", "go")
		b := test.NewStringBuffer()
		s.stderr = b

		s.ctors.NewConfig = func(...config.Option) (*Config, error) {
			return nil, errors.New("config construction failure")
		}

		result := s.Start(args)
		if result != failure {
			t.Errorf("Unexpected result: %d", result)
		}
		if len(b.Buffer) != 1 {
			t.Errorf("Unexpected error log buffer length: %d", len(b.Buffer))
		}
		expected := "Failed to load config: config construction failure\n"
		if !strings.Contains(b.Buffer[0], expected) {
			t.Errorf("\nExpected error log: %s\nActual error log:   %s", expected, b.Buffer[0])
		}
	})

	t.Run("log construction failure", func(t *testing.T) {
		s := New("arcade", "version", "branch", "commit", "date", "go")
		b := test.NewStringBuffer()
		s.stderr = b

		s.ctors.NewConfig = func(...config.Option) (*Config, error) {
			return &Config{Logger: mock.Logger{}}, nil
		}
		s.ctors.NewLogger = func(config.Logger) (log.Logger, error) {
			return nil, errors.New("log construction failure")
		}

		result := s.Start(args)
		if result != failure {
			t.Errorf("Unexpected result: %d", result)
		}
		if len(b.Buffer) != 1 {
			t.Errorf("Unexpected error log buffer length: %d", len(b.Buffer))
		}
		expected := "Failed to create logger: log construction failure\n"
		if !strings.Contains(b.Buffer[0], expected) {
			t.Errorf("\nExpected error log: %s\nActual error log:   %s", expected, b.Buffer[0])
		}
	})

	t.Run("log start", func(t *testing.T) {
		s := New("arcade", "version", "branch", "commit", "date", "go")
		b := test.NewStringBuffer()

		s.ctors.NewConfig = func(...config.Option) (*Config, error) {
			return &Config{Logger: mock.Logger{Level_: "debug", Format_: "logfmt"}}, nil
		}
		s.ctors.NewLogger = func(cfg config.Logger) (log.Logger, error) {
			return log.New(
				log.WithLevel(log.ToLevel(cfg.Level())),
				log.WithFormat(log.ToFormat(cfg.Format())),
				log.WithOutput(b),
			)
		}

		r := make(chan int, 1)
		go func() { r <- s.Start(args) }()
		s.Stop()
		result := <-r

		if result != success {
			t.Errorf("Unexpected result: %d", result)
		}
		if len(b.Buffer) != 1 {
			t.Errorf("Unexpected error log buffer length: %d", len(b.Buffer))
		}
		expected := fmt.Sprintf("level=info msg=starting name=arcade version=version branch=branch commit=commit date=date go=%s", runtime.Version())
		if !strings.Contains(b.Buffer[0], expected) {
			t.Errorf("\nExpected error log: %s\nActual error log:   %s", expected, b.Buffer[0])
		}
	})

	t.Run("success", func(t *testing.T) {
		s := New("arcade", "version", "branch", "commit", "date", "go")

		t.Setenv("ARCADE_POSTGRES_DB", "infra")
		t.Setenv("ARCADE_POSTGRES_HOST", "cockroach")

		r := make(chan int, 1)
		go func() { r <- s.Start(args) }()
		s.Stop()
		result := <-r

		if result != success {
			t.Errorf("Unexpected result: %d", result)
		}
	})
}
