package data_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
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

func TestItemsList(t *testing.T) {
	const (
		cockroachListQ                = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items$"
		cockroachListWithOwnerFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE owner_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithItemFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE location_item_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithPlayerFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE location_player_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithRoomFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE location_room_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithBothFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE owner_id = (.+) AND location_room_id = (.+) LIMIT (.+) OFFSET (.+)$"
	)

	var (
		ctx              = context.Background()
		id               = assets.ItemID(uuid.New())
		name             = "Nothing"
		desc             = "A thing of little importance."
		ownerID          = assets.PlayerID(uuid.New())
		locationItemID   = assets.ItemID(uuid.New())
		locationPlayerID = assets.PlayerID(uuid.New())
		locationRoomID   = assets.RoomID(uuid.New())
		created          = assets.Timestamp{Time: time.Now()}
		updated          = assets.Timestamp{Time: time.Now()}
	)

	t.Run("query error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.ItemDriver
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.ItemDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnError(errors.New("query error"))

			_, err = i.List(ctx, assets.ItemFilter{})

			assert.Error(t, err, `failed to list items: internal server error: query error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("sql scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.ItemDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			_, err = i.List(context.Background(), assets.ItemFilter{})

			assert.Error(t, err, `failed to list items: internal server error: scan error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query      string
			filter     assets.ItemFilter
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID assets.ItemLocationID
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
			},
			{
				query:  cockroachListQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
			},
			{
				query:  cockroachListQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
			{
				query: cockroachListWithOwnerFilterQ,
				filter: assets.ItemFilter{
					OwnerID: ownerID,
					Limit:   assets.DefaultItemFilterLimit,
					Offset:  10,
				},
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
			{
				query: cockroachListWithItemFilterQ,
				filter: assets.ItemFilter{
					LocationID: locationItemID,
					Limit:      assets.DefaultItemFilterLimit,
					Offset:     10,
				},
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
			},
			{
				query: cockroachListWithPlayerFilterQ,
				filter: assets.ItemFilter{
					LocationID: locationPlayerID,
					Limit:      assets.DefaultItemFilterLimit,
					Offset:     10,
				},
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
			},
			{
				query: cockroachListWithRoomFilterQ,
				filter: assets.ItemFilter{
					LocationID: locationRoomID,
					Limit:      assets.DefaultItemFilterLimit,
					Offset:     10,
				},
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
			{
				query: cockroachListWithBothFilterQ,
				filter: assets.ItemFilter{
					OwnerID:    ownerID,
					LocationID: locationRoomID,
					Limit:      assets.DefaultItemFilterLimit,
					Offset:     10,
				},
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			items, err := i.List(context.Background(), test.filter)

			assert.Nil(t, err)
			assert.Equal(t, len(items), 1)
			assert.Compare(t, *items[0], assets.Item{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  test.locationID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestItemsGet(t *testing.T) {
	const (
		cockroachGetQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items WHERE id = (.+)$"
	)

	var (
		ctx              = context.Background()
		id               = assets.ItemID(uuid.New())
		name             = "Nothing"
		desc             = "A thing of little importance."
		ownerID          = assets.PlayerID(uuid.New())
		locationItemID   = assets.ItemID(uuid.New())
		locationPlayerID = assets.PlayerID(uuid.New())
		locationRoomID   = assets.RoomID(uuid.New())
		created          = assets.Timestamp{Time: time.Now()}
		updated          = assets.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.ItemDriver
			err    error
			msg    string
		}{
			{
				query:  cockroachGetQ,
				driver: cockroach.ItemDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to get item: not found",
			},
			{
				query:  cockroachGetQ,
				driver: cockroach.ItemDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to get item: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WithArgs(id).WillReturnError(test.err)

			_, err = i.Get(ctx, id)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID assets.ItemLocationID
		}{
			{
				query:  cockroachGetQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
			},
			{
				query:  cockroachGetQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
			},
			{
				query:  cockroachGetQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows)

			item, err := i.Get(ctx, id)

			assert.Nil(t, err)
			assert.Compare(t, *item, assets.Item{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  test.locationID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestItemsCreate(t *testing.T) {
	const (
		cockroachCreateQ = `^INSERT INTO items \(name, description, owner_id, location_item_id, location_player_id, location_room_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated$`
	)

	var (
		ctx              = context.Background()
		id               = assets.ItemID(uuid.New())
		name             = "Nothing"
		desc             = "A thing of little importance."
		ownerID          = assets.PlayerID(uuid.New())
		locationItemID   = assets.ItemID(uuid.New())
		locationPlayerID = assets.PlayerID(uuid.New())
		locationRoomID   = assets.RoomID(uuid.New())
		created          = assets.Timestamp{Time: time.Now()}
		updated          = assets.Timestamp{Time: time.Now()}
	)

	t.Run("violations", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID assets.ItemLocationID
			args       []driver.Value
			err        error
			msg        string
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
				err:        &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to create item: bad request: the given ownerID or locationID does not exist: "+
					"ownerID '%s', locationID '%s (%s)'", ownerID, locationItemID, "item"),
			},
			{
				query:  cockroachCreateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
				args:       []driver.Value{uuid.NullUUID{}, locationPlayerID, uuid.NullUUID{}},
				err:        &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg:        "failed to create item: bad request: item name 'Nothing' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			create := assets.ItemCreate{
				ItemChange: assets.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, test.args[0], test.args[1], test.args[2]).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = i.Create(ctx, create)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID assets.ItemLocationID
			args       []driver.Value
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated).
					RowError(0, errors.New("scan error")),
				locationID: locationRoomID,
				args:       []driver.Value{uuid.NullUUID{}, uuid.NullUUID{}, locationRoomID},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			create := assets.ItemCreate{
				ItemChange: assets.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, test.args[0], test.args[1], test.args[2]).WillReturnRows(test.rows)

			_, err = i.Create(ctx, create)

			assert.Error(t, err, "failed to create item: internal server error: scan error")
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID assets.ItemLocationID
			args       []driver.Value
		}{
			{
				query:  cockroachCreateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
			},
			{
				query:  cockroachCreateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
				args:       []driver.Value{uuid.NullUUID{}, locationPlayerID, uuid.NullUUID{}},
			},
			{
				query:  cockroachCreateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
				args:       []driver.Value{uuid.NullUUID{}, uuid.NullUUID{}, locationRoomID},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}

			create := assets.ItemCreate{
				ItemChange: assets.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, test.args[0], test.args[1], test.args[2]).WillReturnRows(test.rows)

			item, err := i.Create(ctx, create)

			assert.Nil(t, err)
			assert.Compare(t, *item, assets.Item{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  test.locationID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestItemsUpdate(t *testing.T) {
	const (
		cockroachUpdateQ = `^UPDATE items SET name = (.+), description = (.+), owner_id = (.+), location_item_id = (.+), location_player_id = (.+), location_room_id = (.+) ` +
			`WHERE id = (.+) ` +
			`RETURNING id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated$`
	)

	var (
		ctx              = context.Background()
		id               = assets.ItemID(uuid.New())
		name             = "Nothing"
		desc             = "A thing of little importance."
		ownerID          = assets.PlayerID(uuid.New())
		locationItemID   = assets.ItemID(uuid.New())
		locationPlayerID = assets.PlayerID(uuid.New())
		locationRoomID   = assets.RoomID(uuid.New())
		created          = assets.Timestamp{Time: time.Now()}
		updated          = assets.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID assets.ItemLocationID
			args       []driver.Value
			err        error
			msg        string
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
				err:        sql.ErrNoRows,
				msg:        "failed to update item: not found",
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
				args:       []driver.Value{uuid.NullUUID{}, locationPlayerID, uuid.NullUUID{}},
				err:        &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to update item: bad request: the given ownerID or locationID does not exist: "+
					"ownerID '%s', locationID '%s (%s)'", ownerID, locationPlayerID, "player"),
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
				args:       []driver.Value{uuid.NullUUID{}, uuid.NullUUID{}, locationRoomID},
				err:        &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg:        "failed to update item: bad request: item name 'Nothing' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			update := assets.ItemUpdate{
				ItemChange: assets.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, test.args[0], test.args[1], test.args[2]).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = i.Update(ctx, id, update)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID assets.ItemLocationID
			args       []driver.Value
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated).
					RowError(0, errors.New("scan error")),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			update := assets.ItemUpdate{
				ItemChange: assets.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, test.args[0], test.args[1], test.args[2]).WillReturnRows(test.rows)

			_, err = i.Update(ctx, id, update)

			assert.Error(t, err, "failed to update item: internal server error: scan error")
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID assets.ItemLocationID
			args       []driver.Value
		}{
			{
				query:  cockroachUpdateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
				args:       []driver.Value{uuid.NullUUID{}, locationPlayerID, uuid.NullUUID{}},
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
				args:       []driver.Value{uuid.NullUUID{}, uuid.NullUUID{}, locationRoomID},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			update := assets.ItemUpdate{
				ItemChange: assets.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, test.args[0], test.args[1], test.args[2]).WillReturnRows(test.rows)

			item, err := i.Update(ctx, id, update)

			assert.Nil(t, err)
			assert.Compare(t, *item, assets.Item{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  test.locationID,
				Created:     created,
				Updated:     updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestItemsRemove(t *testing.T) {
	const (
		cockroachRemoveQ = `^DELETE FROM items WHERE id = (.+)$`
	)

	var (
		ctx = context.Background()
		id  = assets.ItemID(uuid.New())
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.ItemDriver
			err    error
			msg    string
		}{
			{
				query:  cockroachRemoveQ,
				driver: cockroach.ItemDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to remove item: not found",
			},
			{
				query:  cockroachRemoveQ,
				driver: cockroach.ItemDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to remove item: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnError(test.err)

			err = i.Remove(ctx, id)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.ItemDriver
		}{
			{
				query:  cockroachRemoveQ,
				driver: cockroach.ItemDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: db, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

			err = i.Remove(ctx, id)

			assert.Nil(t, err)
			assert.MockExpectationsMet(t, mock)
		}
	})
}
