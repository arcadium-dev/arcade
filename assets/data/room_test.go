package data_test

import (
	"context"
	"database/sql"
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

	"arcadium.dev/arcade/assets"
	"arcadium.dev/arcade/assets/data"
	"arcadium.dev/arcade/assets/data/cockroach"
)

func TestRoomsList(t *testing.T) {
	const (
		cockroachListQ               = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms$"
		cockroachListWithBothFilterQ = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms" +
			" WHERE owner_id = (.+) AND parent_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithOwnerFilterQ = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms" +
			" WHERE owner_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithParentFilterQ = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms" +
			" WHERE parent_id = (.+) LIMIT (.+) OFFSET (.+)$"
	)

	var (
		ctx      = context.Background()
		id       = assets.RoomID(uuid.New())
		name     = "Nobody"
		desc     = "No one of importance."
		ownerID  = assets.PlayerID(uuid.New())
		parentID = assets.RoomID(uuid.New())
		created  = assets.Timestamp{Time: time.Now()}
		updated  = assets.Timestamp{Time: time.Now()}
	)

	t.Run("query error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.RoomDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnError(errors.New("query error"))

			_, err = p.List(ctx, assets.RoomFilter{})

			assert.Error(t, err, `failed to list rooms: internal server error: query error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("sql scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			_, err = p.List(context.Background(), assets.RoomFilter{})

			assert.Error(t, err, `failed to list rooms: internal server error: scan error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			filter assets.RoomFilter
			driver data.RoomDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
			{
				query: cockroachListWithBothFilterQ,
				filter: assets.RoomFilter{
					OwnerID:  ownerID,
					ParentID: parentID,
					Limit:    assets.DefaultRoomFilterLimit,
					Offset:   10,
				},
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
			{
				query: cockroachListWithOwnerFilterQ,
				filter: assets.RoomFilter{
					OwnerID: ownerID,
					Limit:   assets.DefaultRoomFilterLimit,
					Offset:  10,
				},
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
			{
				query: cockroachListWithParentFilterQ,
				filter: assets.RoomFilter{
					ParentID: parentID,
					Limit:    assets.DefaultRoomFilterLimit,
					Offset:   10,
				},
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			rooms, err := p.List(context.Background(), test.filter)

			assert.Nil(t, err)
			assert.Equal(t, len(rooms), 1)
			assert.Compare(t, *rooms[0], assets.Room{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestRoomsGet(t *testing.T) {
	const (
		cockroachGetQ = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms WHERE id = (.+)$"
	)

	var (
		ctx      = context.Background()
		id       = assets.RoomID(uuid.New())
		name     = "Nobody"
		desc     = "No one of importance."
		ownerID  = assets.PlayerID(uuid.New())
		parentID = assets.RoomID(uuid.New())
		created  = assets.Timestamp{Time: time.Now()}
		updated  = assets.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			err    error
			msg    string
		}{
			{
				query:  cockroachGetQ,
				driver: cockroach.RoomDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to get room: not found",
			},
			{
				query:  cockroachGetQ,
				driver: cockroach.RoomDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to get room: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WithArgs(id).WillReturnError(test.err)

			_, err = p.Get(ctx, id)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachGetQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows)

			room, err := p.Get(ctx, id)

			assert.Nil(t, err)
			assert.Compare(t, *room, assets.Room{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestRoomsCreate(t *testing.T) {
	const (
		cockroachCreateQ = `^INSERT INTO rooms \(name, description, owner_id, parent_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING id, name, description, owner_id, parent_id, created, updated$`
	)

	var (
		ctx      = context.Background()
		id       = assets.RoomID(uuid.New())
		name     = "Nobody"
		desc     = "No one of importance."
		ownerID  = assets.PlayerID(uuid.New())
		parentID = assets.RoomID(uuid.New())
		created  = assets.Timestamp{Time: time.Now()}
		updated  = assets.Timestamp{Time: time.Now()}
	)

	t.Run("violations", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			rows   *sqlmock.Rows
			err    error
			msg    string
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to create room: bad request: the given ownerID or parentID does not exist: "+
					"ownerID '%s', parentID '%s'", ownerID, parentID),
			},
			{
				query:  cockroachCreateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg: "failed to create room: bad request: room name 'Nobody' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			create := assets.RoomCreate{
				RoomChange: assets.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, parentID).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = p.Create(ctx, create)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			create := assets.RoomCreate{
				RoomChange: assets.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, parentID).WillReturnRows(test.rows)

			_, err = p.Create(ctx, create)

			assert.Error(t, err, "failed to create room: internal server error: scan error")
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}

			create := assets.RoomCreate{
				RoomChange: assets.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, parentID).WillReturnRows(test.rows)

			room, err := p.Create(ctx, create)

			assert.Nil(t, err)
			assert.Compare(t, *room, assets.Room{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestRoomsUpdate(t *testing.T) {
	const (
		cockroachUpdateQ = `^UPDATE rooms SET name = (.+), description = (.+), owner_id = (.+), parent_id = (.+) ` +
			`WHERE id = (.+) ` +
			`RETURNING id, name, description, owner_id, parent_id, created, updated$`
	)

	var (
		ctx      = context.Background()
		id       = assets.RoomID(uuid.New())
		name     = "Nobody"
		desc     = "No one of importance."
		ownerID  = assets.PlayerID(uuid.New())
		parentID = assets.RoomID(uuid.New())
		created  = assets.Timestamp{Time: time.Now()}
		updated  = assets.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			rows   *sqlmock.Rows
			err    error
			msg    string
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: sql.ErrNoRows,
				msg: "failed to update room: not found",
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to update room: bad request: the given ownerID or parentID does not exist: "+
					"ownerID '%s', parentID '%s'", ownerID, parentID),
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg: "failed to update room: bad request: room name 'Nobody' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			update := assets.RoomUpdate{
				RoomChange: assets.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, parentID).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = p.Update(ctx, id, update)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			update := assets.RoomUpdate{
				RoomChange: assets.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, parentID).WillReturnRows(test.rows)

			_, err = p.Update(ctx, id, update)

			assert.Error(t, err, "failed to update room: internal server error: scan error")
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			update := assets.RoomUpdate{
				RoomChange: assets.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, parentID).WillReturnRows(test.rows)

			room, err := p.Update(ctx, id, update)

			assert.Nil(t, err)
			assert.Compare(t, *room, assets.Room{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestRoomsRemove(t *testing.T) {
	const (
		cockroachRemoveQ = `^DELETE FROM rooms WHERE id = (.+)$`
	)

	var (
		ctx = context.Background()
		id  = assets.RoomID(uuid.New())
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			err    error
			msg    string
		}{
			{
				query:  cockroachRemoveQ,
				driver: cockroach.RoomDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to remove room: not found",
			},
			{
				query:  cockroachRemoveQ,
				driver: cockroach.RoomDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to remove room: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnError(test.err)

			err = p.Remove(ctx, id)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
		}{
			{
				query:  cockroachRemoveQ,
				driver: cockroach.RoomDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			p := data.RoomStorage{DB: db, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

			err = p.Remove(ctx, id)

			assert.Nil(t, err)
			assert.MockExpectationsMet(t, mock)
		}
	})
}
