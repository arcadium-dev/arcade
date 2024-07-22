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
	"arcadium.dev/arcade/asset/data/postgres"
)

func TestRoomsList(t *testing.T) {
	const (
		postgresListQ               = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms$"
		postgresListWithBothFilterQ = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms" +
			" WHERE owner_id = (.+) AND parent_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithOwnerFilterQ = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms" +
			" WHERE owner_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithParentFilterQ = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms" +
			" WHERE parent_id = (.+) LIMIT (.+) OFFSET (.+)$"
	)

	var (
		ctx      = context.Background()
		id       = asset.RoomID(uuid.New())
		name     = "Nowhere"
		desc     = "A place of no importance."
		ownerID  = asset.PlayerID(uuid.New())
		parentID = asset.RoomID(uuid.New())
		created  = asset.Timestamp{Time: time.Now()}
		updated  = asset.Timestamp{Time: time.Now()}
	)

	t.Run("query error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
		}{
			{
				query:  postgresListQ,
				driver: postgres.RoomDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnError(errors.New("query error"))

			_, err = r.List(ctx, asset.RoomFilter{})

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
				query:  postgresListQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			_, err = r.List(context.Background(), asset.RoomFilter{})

			assert.Error(t, err, `failed to list rooms: internal server error: scan error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			filter asset.RoomFilter
			driver data.RoomDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  postgresListQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
			{
				query: postgresListWithBothFilterQ,
				filter: asset.RoomFilter{
					OwnerID:  ownerID,
					ParentID: parentID,
					Limit:    asset.DefaultRoomFilterLimit,
					Offset:   10,
				},
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
			{
				query: postgresListWithOwnerFilterQ,
				filter: asset.RoomFilter{
					OwnerID: ownerID,
					Limit:   asset.DefaultRoomFilterLimit,
					Offset:  10,
				},
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
			{
				query: postgresListWithParentFilterQ,
				filter: asset.RoomFilter{
					ParentID: parentID,
					Limit:    asset.DefaultRoomFilterLimit,
					Offset:   10,
				},
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			rooms, err := r.List(context.Background(), test.filter)

			assert.Nil(t, err)
			assert.Equal(t, len(rooms), 1)
			assert.Compare(t, *rooms[0], asset.Room{
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
		postgresGetQ = "^SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms WHERE id = (.+)$"
	)

	var (
		ctx      = context.Background()
		id       = asset.RoomID(uuid.New())
		name     = "Nowhere"
		desc     = "A place of no importance."
		ownerID  = asset.PlayerID(uuid.New())
		parentID = asset.RoomID(uuid.New())
		created  = asset.Timestamp{Time: time.Now()}
		updated  = asset.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			err    error
			msg    string
		}{
			{
				query:  postgresGetQ,
				driver: postgres.RoomDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to get room: not found",
			},
			{
				query:  postgresGetQ,
				driver: postgres.RoomDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to get room: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WithArgs(id).WillReturnError(test.err)

			_, err = r.Get(ctx, id)

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
				query:  postgresGetQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows)

			room, err := r.Get(ctx, id)

			assert.Nil(t, err)
			assert.Compare(t, *room, asset.Room{
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
		postgresCreateQ = `^INSERT INTO rooms \(name, description, owner_id, parent_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING id, name, description, owner_id, parent_id, created, updated$`
	)

	var (
		ctx      = context.Background()
		id       = asset.RoomID(uuid.New())
		name     = "Nowhere"
		desc     = "A place of no importance."
		ownerID  = asset.PlayerID(uuid.New())
		parentID = asset.RoomID(uuid.New())
		created  = asset.Timestamp{Time: time.Now()}
		updated  = asset.Timestamp{Time: time.Now()}
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
				query:  postgresCreateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to create room: bad request: the given ownerID or parentID does not exist: "+
					"ownerID '%s', parentID '%s'", ownerID, parentID),
			},
			{
				query:  postgresCreateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg: "failed to create room: bad request: room name 'Nowhere' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := asset.RoomCreate{
				RoomChange: asset.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, parentID).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = r.Create(ctx, create)

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
				query:  postgresCreateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := asset.RoomCreate{
				RoomChange: asset.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, parentID).WillReturnRows(test.rows)

			_, err = r.Create(ctx, create)

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
				query:  postgresCreateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}

			create := asset.RoomCreate{
				RoomChange: asset.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, parentID).WillReturnRows(test.rows)

			room, err := r.Create(ctx, create)

			assert.Nil(t, err)
			assert.Compare(t, *room, asset.Room{
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
		postgresUpdateQ = `^UPDATE rooms SET name = (.+), description = (.+), owner_id = (.+), parent_id = (.+) ` +
			`WHERE id = (.+) ` +
			`RETURNING id, name, description, owner_id, parent_id, created, updated$`
	)

	var (
		ctx      = context.Background()
		id       = asset.RoomID(uuid.New())
		name     = "Nowhere"
		desc     = "A place of no importance."
		ownerID  = asset.PlayerID(uuid.New())
		parentID = asset.RoomID(uuid.New())
		created  = asset.Timestamp{Time: time.Now()}
		updated  = asset.Timestamp{Time: time.Now()}
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
				query:  postgresUpdateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: sql.ErrNoRows,
				msg: "failed to update room: not found",
			},
			{
				query:  postgresUpdateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to update room: bad request: the given ownerID or parentID does not exist: "+
					"ownerID '%s', parentID '%s'", ownerID, parentID),
			},
			{
				query:  postgresUpdateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg: "failed to update room: bad request: room name 'Nowhere' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.RoomUpdate{
				RoomChange: asset.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, parentID).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = r.Update(ctx, id, update)

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
				query:  postgresUpdateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.RoomUpdate{
				RoomChange: asset.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, parentID).WillReturnRows(test.rows)

			_, err = r.Update(ctx, id, update)

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
				query:  postgresUpdateQ,
				driver: postgres.RoomDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, parentID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.RoomUpdate{
				RoomChange: asset.RoomChange{Name: name, Description: desc, OwnerID: ownerID, ParentID: parentID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, parentID).WillReturnRows(test.rows)

			room, err := r.Update(ctx, id, update)

			assert.Nil(t, err)
			assert.Compare(t, *room, asset.Room{
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
		postgresRemoveQ = `^DELETE FROM rooms WHERE id = (.+)$`
	)

	var (
		ctx = context.Background()
		id  = asset.RoomID(uuid.New())
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.RoomDriver
			err    error
			msg    string
		}{
			{
				query:  postgresRemoveQ,
				driver: postgres.RoomDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to remove room: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnError(test.err)

			err = r.Remove(ctx, id)

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
				query:  postgresRemoveQ,
				driver: postgres.RoomDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			r := data.RoomStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

			err = r.Remove(ctx, id)

			assert.Nil(t, err)
			assert.MockExpectationsMet(t, mock)
		}
	})
}
