package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"arcadium.dev/core/sql"

	pg "arcadium.dev/arcade/asset/data/postgres"
)

var (
	db *sql.DB
)

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Hostname:   "postgres-dockertest",
		Name:       "postgres-dockertest",
		Repository: "postgres",
		Tag:        "16",
		Env: []string{
			"POSTGRES_PASSWORD=arcadium",
			"POSTGRES_USER=arcadium",
			"POSTGRES_DB=arcade",
			"POSTGRES_HOST_AUTH_METHOD=trust",
			"listen_addresses = '*'",
		},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432/tcp": {{HostPort: "5432"}},
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	dsn := "postgres://arcadium:arcadium@localhost:5432/arcade?sslmode=disable"

	log.Printf("Connecting to database on url: %s", dsn)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		ctx := context.Background()

		log.Println("Trying to connect...")
		db, err = pg.Open(ctx, dsn)
		if err != nil {
			log.Printf("Failed to connect, '%s'", err)
			return fmt.Errorf("failed to open database: %w", err)
		}

		err = db.Ping(ctx)
		if err != nil {
			log.Printf("Ping failed, '%s'", err)
		}
		log.Println("... connected.")
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}()

	// Migrate DB.
	mig, err := newMigration(db)
	if err != nil {
		log.Fatalf("Could not create migrate: %s", err)
	}

	if err := mig.Up(); err != nil {
		log.Fatalf("Could not migrate up: %s", err)
	}

	// Tell docker to hard kill the container 10 seconds from now.
	if err := resource.Expire(10); err != nil {
		log.Fatalf("Cound not set resource expiry")
	}

	// run tests
	code := m.Run()

	if err := mig.Down(); err != nil {
		log.Fatalf("Could not migrate down: %s", err)
	}

	// You can't defer this because os.Exit doesn't care for defer.
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

type (
	migration struct {
		m *migrate.Migrate
	}
)

const (
	migrationsPath = "./migrations"
)

func newMigration(db *sql.DB) (*migration, error) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("")
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "postgress", driver)
	if err != nil {
		return nil, err
	}
	return &migration{m: m}, nil
}

func (m *migration) Up() error {
	log.Printf("migrate up")
	err := m.m.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
		log.Printf("error: %s", err)
		return nil
	case err != nil:
		return err
	}
	return nil
}

func (m *migration) Down() error {
	log.Printf("migrate down")
	err := m.m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
