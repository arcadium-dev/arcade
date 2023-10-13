package main_test

/*
import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/http/server"

	cmd "arcadium.dev/arcade/cmd/assets"
)

// Note: TestMain is a low-level primitive and can cause issues with IDEs
func TestMainEntryPoint(t *testing.T) {
	t.Setenv("DSN", "foobar")

	t.Run("init failure", func(t *testing.T) {
		restore := setNew(func(v, b, c, d string) cmd.RestServer {
			return mockRestServer{initErr: errors.New("init failure")}
		})
		defer restore()

		err := cmd.Main()

		assert.Error(t, err, "init failure")
	})

	t.Run("start failure", func(t *testing.T) {
		restore := setNew(func(v, b, c, d string) cmd.RestServer {
			return mockRestServer{startErr: errors.New("start failure")}
		})
		defer restore()

		err := cmd.Main()

		assert.Error(t, err, "start failure")
	})

	t.Run("success", func(t *testing.T) {
		restore := setNew(func(v, b, c, d string) cmd.RestServer {
			return mockRestServer{}
		})
		defer restore()

		err := cmd.Main()

		assert.Nil(t, err)
	})
}

func setNew(newNew func(v, b, c, d string) cmd.RestServer) func() {
	oldNew := cmd.New
	cmd.New = newNew
	return func() { cmd.New = oldNew }
}

type (
	mockRestServer struct {
		db                *sql.DB
		initErr, startErr error
	}
)

func (m mockRestServer) Init(...string) error {
	return m.initErr
}

func (m mockRestServer) Start(...server.Service) error {
	return m.startErr
}

func (m mockRestServer) DB() *sql.DB {
	return m.db
}

func (m mockRestServer) Ctx() context.Context {
	return context.Background()
}
*/
