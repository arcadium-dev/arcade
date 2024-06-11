package data_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/data"
	"arcadium.dev/arcade/asset/data/cockroach"
)

func TestPlayersList(t *testing.T) {
	const (
		cockroachListQ           = "^SELECT id, name, description, home_id, location_id, created, updated FROM players$"
		cockroachListWithFilterQ = "^SELECT id, name, description, home_id, location_id, created, updated FROM players" +
			" WHERE location_id = (.+) LIMIT (.+) OFFSET (.+)$"
	)

	var (
		ctx        = context.Background()
		id         = asset.PlayerID(uuid.New())
		name       = "Nobody"
		desc       = "No one of importance."
		homeID     = asset.RoomID(uuid.New())
		locationID = asset.RoomID(uuid.New())
		created    = asset.Timestamp{Time: time.Now()}
		updated    = asset.Timestamp{Time: time.Now()}
	)

	t.Run("query error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.PlayerDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnError(errors.New("query error"))

			_, err = p.List(ctx, asset.PlayerFilter{})

			assert.Error(t, err, `failed to list players: internal server error: query error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("sql scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			_, err = p.List(context.Background(), asset.PlayerFilter{})

			assert.Error(t, err, `failed to list players: internal server error: scan error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			filter asset.PlayerFilter
			driver data.PlayerDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
			},
			{
				query: cockroachListWithFilterQ,
				filter: asset.PlayerFilter{
					LocationID: locationID,
					Limit:      asset.DefaultPlayerFilterLimit,
					Offset:     10,
				},
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			players, err := p.List(context.Background(), test.filter)

			assert.Nil(t, err)
			assert.Equal(t, len(players), 1)
			assert.Compare(t, *players[0], asset.Player{
				ID:          id,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestPlayersGet(t *testing.T) {
	const (
		cockroachGetQ = "^SELECT id, name, description, home_id, location_id, created, updated FROM players WHERE id = (.+)$"
	)

	var (
		ctx        = context.Background()
		id         = asset.PlayerID(uuid.New())
		name       = "Nobody"
		desc       = "No one of importance."
		homeID     = asset.RoomID(uuid.New())
		locationID = asset.RoomID(uuid.New())
		created    = asset.Timestamp{Time: time.Now()}
		updated    = asset.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			err    error
			msg    string
		}{
			{
				query:  cockroachGetQ,
				driver: cockroach.PlayerDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to get player: not found",
			},
			{
				query:  cockroachGetQ,
				driver: cockroach.PlayerDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to get player: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WithArgs(id).WillReturnError(test.err)

			_, err = p.Get(ctx, id)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachGetQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows)

			player, err := p.Get(ctx, id)

			assert.Nil(t, err)
			assert.Compare(t, *player, asset.Player{
				ID:          id,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestPlayersCreate(t *testing.T) {
	const (
		cockroachCreateQ = `^INSERT INTO players \(name, description, home_id, location_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING id, name, description, home_id, location_id, created, updated$`
	)

	var (
		ctx        = context.Background()
		id         = asset.PlayerID(uuid.New())
		name       = "Nobody"
		desc       = "No one of importance."
		homeID     = asset.RoomID(uuid.New())
		locationID = asset.RoomID(uuid.New())
		created    = asset.Timestamp{Time: time.Now()}
		updated    = asset.Timestamp{Time: time.Now()}
	)

	t.Run("violations", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			rows   *sqlmock.Rows
			err    error
			msg    string
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to create player: bad request: the given homeID or locationID does not exist: "+
					"homeID '%s', locationID '%s'", homeID, locationID),
			},
			{
				query:  cockroachCreateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg: "failed to create player: bad request: player name 'Nobody' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := asset.PlayerCreate{
				PlayerChange: asset.PlayerChange{Name: name, Description: desc, HomeID: homeID, LocationID: locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, homeID, locationID).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = p.Create(ctx, create)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := asset.PlayerCreate{
				PlayerChange: asset.PlayerChange{Name: name, Description: desc, HomeID: homeID, LocationID: locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, homeID, locationID).WillReturnRows(test.rows)

			_, err = p.Create(ctx, create)

			assert.Error(t, err, "failed to create player: internal server error: scan error")
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}

			create := asset.PlayerCreate{
				PlayerChange: asset.PlayerChange{Name: name, Description: desc, HomeID: homeID, LocationID: locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, homeID, locationID).WillReturnRows(test.rows)

			player, err := p.Create(ctx, create)

			assert.Nil(t, err)
			assert.Compare(t, *player, asset.Player{
				ID:          id,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestPlayersUpdate(t *testing.T) {
	const (
		cockroachUpdateQ = `^UPDATE players SET name = (.+), description = (.+), home_id = (.+), location_id = (.+) ` +
			`WHERE id = (.+) ` +
			`RETURNING id, name, description, home_id, location_id, created, updated$`
	)

	var (
		ctx        = context.Background()
		id         = asset.PlayerID(uuid.New())
		name       = "Nobody"
		desc       = "No one of importance."
		homeID     = asset.RoomID(uuid.New())
		locationID = asset.RoomID(uuid.New())
		created    = asset.Timestamp{Time: time.Now()}
		updated    = asset.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			rows   *sqlmock.Rows
			err    error
			msg    string
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
				err: sql.ErrNoRows,
				msg: "failed to update player: not found",
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to update player: bad request: the given homeID or locationID does not exist: "+
					"homeID '%s', locationID '%s'", homeID, locationID),
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg: "failed to update player: bad request: player name 'Nobody' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.PlayerUpdate{
				PlayerChange: asset.PlayerChange{Name: name, Description: desc, HomeID: homeID, LocationID: locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, homeID, locationID).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = p.Update(ctx, id, update)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.PlayerUpdate{
				PlayerChange: asset.PlayerChange{Name: name, Description: desc, HomeID: homeID, LocationID: locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, homeID, locationID).WillReturnRows(test.rows)

			_, err = p.Update(ctx, id, update)

			assert.Error(t, err, "failed to update player: internal server error: scan error")
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.PlayerDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "home_id", "location_id", "created", "updated"}).
					AddRow(id, name, desc, homeID, locationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.PlayerUpdate{
				PlayerChange: asset.PlayerChange{Name: name, Description: desc, HomeID: homeID, LocationID: locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, homeID, locationID).WillReturnRows(test.rows)

			player, err := p.Update(ctx, id, update)

			assert.Nil(t, err)
			assert.Compare(t, *player, asset.Player{
				ID:          id,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestPlayersRemove(t *testing.T) {
	const (
		cockroachRemoveQ = `^DELETE FROM players WHERE id = (.+)$`
	)

	var (
		ctx = context.Background()
		id  = asset.PlayerID(uuid.New())
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
			err    error
			msg    string
		}{
			{
				query:  cockroachRemoveQ,
				driver: cockroach.PlayerDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to remove player: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnError(test.err)

			err = p.Remove(ctx, id)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.PlayerDriver
		}{
			{
				query:  cockroachRemoveQ,
				driver: cockroach.PlayerDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.PlayerStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

			err = p.Remove(ctx, id)

			assert.Nil(t, err)
			assert.MockExpectationsMet(t, mock)
		}
	})
}
