package data_test

import (
	"context"
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
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/data"
	"arcadium.dev/arcade/asset/data/postgres"
)

func TestItemsList(t *testing.T) {
	const (
		postgresListQ                = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items$"
		postgresListWithOwnerFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE owner_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithItemFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE location_item_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithPlayerFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE location_player_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithRoomFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE location_room_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithBothFilterQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items" +
			" WHERE owner_id = (.+) AND location_room_id = (.+) LIMIT (.+) OFFSET (.+)$"
	)

	var (
		ctx              = context.Background()
		id               = asset.ItemID(uuid.New())
		name             = "Nothing"
		desc             = "A thing of little importance."
		ownerID          = asset.PlayerID(uuid.New())
		locationItemID   = asset.ItemID(uuid.New())
		locationPlayerID = asset.PlayerID(uuid.New())
		locationRoomID   = asset.RoomID(uuid.New())
		created          = arcade.Timestamp{Time: time.Now()}
		updated          = arcade.Timestamp{Time: time.Now()}
	)

	t.Run("query error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.ItemDriver
		}{
			{
				query:  postgresListQ,
				driver: postgres.ItemDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnError(errors.New("query error"))

			_, err = i.List(ctx, asset.ItemFilter{})

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
				query:  postgresListQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			_, err = i.List(context.Background(), asset.ItemFilter{})

			assert.Error(t, err, `failed to list items: internal server error: scan error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query      string
			filter     asset.ItemFilter
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID asset.ItemLocationID
		}{
			{
				query:  postgresListQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
			},
			{
				query:  postgresListQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
			},
			{
				query:  postgresListQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
			{
				query: postgresListWithOwnerFilterQ,
				filter: asset.ItemFilter{
					OwnerID: ownerID,
					Limit:   asset.DefaultItemFilterLimit,
					Offset:  10,
				},
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
			{
				query: postgresListWithItemFilterQ,
				filter: asset.ItemFilter{
					LocationID: locationItemID,
					Limit:      asset.DefaultItemFilterLimit,
					Offset:     10,
				},
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
			},
			{
				query: postgresListWithPlayerFilterQ,
				filter: asset.ItemFilter{
					LocationID: locationPlayerID,
					Limit:      asset.DefaultItemFilterLimit,
					Offset:     10,
				},
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
			},
			{
				query: postgresListWithRoomFilterQ,
				filter: asset.ItemFilter{
					LocationID: locationRoomID,
					Limit:      asset.DefaultItemFilterLimit,
					Offset:     10,
				},
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
			{
				query: postgresListWithBothFilterQ,
				filter: asset.ItemFilter{
					OwnerID:    ownerID,
					LocationID: locationRoomID,
					Limit:      asset.DefaultItemFilterLimit,
					Offset:     10,
				},
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			items, err := i.List(context.Background(), test.filter)

			assert.Nil(t, err)
			assert.Equal(t, len(items), 1)
			assert.Compare(t, *items[0], asset.Item{
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
		postgresGetQ = "^SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items WHERE id = (.+)$"
	)

	var (
		ctx              = context.Background()
		id               = asset.ItemID(uuid.New())
		name             = "Nothing"
		desc             = "A thing of little importance."
		ownerID          = asset.PlayerID(uuid.New())
		locationItemID   = asset.ItemID(uuid.New())
		locationPlayerID = asset.PlayerID(uuid.New())
		locationRoomID   = asset.RoomID(uuid.New())
		created          = arcade.Timestamp{Time: time.Now()}
		updated          = arcade.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.ItemDriver
			err    error
			msg    string
		}{
			{
				query:  postgresGetQ,
				driver: postgres.ItemDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to get item: not found",
			},
			{
				query:  postgresGetQ,
				driver: postgres.ItemDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to get item: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
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
			locationID asset.ItemLocationID
		}{
			{
				query:  postgresGetQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
			},
			{
				query:  postgresGetQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
			},
			{
				query:  postgresGetQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows)

			item, err := i.Get(ctx, id)

			assert.Nil(t, err)
			assert.Compare(t, *item, asset.Item{
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
		postgresCreateQ = `^INSERT INTO items \(name, description, owner_id, location_item_id, location_player_id, location_room_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated$`
	)

	var (
		ctx              = context.Background()
		id               = asset.ItemID(uuid.New())
		name             = "Nothing"
		desc             = "A thing of little importance."
		ownerID          = asset.PlayerID(uuid.New())
		locationItemID   = asset.ItemID(uuid.New())
		locationPlayerID = asset.PlayerID(uuid.New())
		locationRoomID   = asset.RoomID(uuid.New())
		created          = arcade.Timestamp{Time: time.Now()}
		updated          = arcade.Timestamp{Time: time.Now()}
	)

	t.Run("violations", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID asset.ItemLocationID
			args       []driver.Value
			err        error
			msg        string
		}{
			{
				query:  postgresCreateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
				err:        &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to create item: bad request: the given ownerID or locationID does not exist: "+
					"ownerID '%s', locationID '%s (%s)'", ownerID, locationItemID, "item"),
			},
			{
				query:  postgresCreateQ,
				driver: postgres.ItemDriver{},
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

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := asset.ItemCreate{
				ItemChange: asset.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
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
			locationID asset.ItemLocationID
			args       []driver.Value
		}{
			{
				query:  postgresCreateQ,
				driver: postgres.ItemDriver{},
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

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := asset.ItemCreate{
				ItemChange: asset.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
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
			locationID asset.ItemLocationID
			args       []driver.Value
		}{
			{
				query:  postgresCreateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
			},
			{
				query:  postgresCreateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
				args:       []driver.Value{uuid.NullUUID{}, locationPlayerID, uuid.NullUUID{}},
			},
			{
				query:  postgresCreateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
				args:       []driver.Value{uuid.NullUUID{}, uuid.NullUUID{}, locationRoomID},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}

			create := asset.ItemCreate{
				ItemChange: asset.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, test.args[0], test.args[1], test.args[2]).WillReturnRows(test.rows)

			item, err := i.Create(ctx, create)

			assert.Nil(t, err)
			assert.Compare(t, *item, asset.Item{
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
		postgresUpdateQ = `^UPDATE items SET name = (.+), description = (.+), owner_id = (.+), location_item_id = (.+), location_player_id = (.+), location_room_id = (.+) ` +
			`WHERE id = (.+) ` +
			`RETURNING id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated$`
	)

	var (
		ctx              = context.Background()
		id               = asset.ItemID(uuid.New())
		name             = "Nothing"
		desc             = "A thing of little importance."
		ownerID          = asset.PlayerID(uuid.New())
		locationItemID   = asset.ItemID(uuid.New())
		locationPlayerID = asset.PlayerID(uuid.New())
		locationRoomID   = asset.RoomID(uuid.New())
		created          = arcade.Timestamp{Time: time.Now()}
		updated          = arcade.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query      string
			driver     data.ItemDriver
			rows       *sqlmock.Rows
			locationID asset.ItemLocationID
			args       []driver.Value
			err        error
			msg        string
		}{
			{
				query:  postgresUpdateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
				err:        sql.ErrNoRows,
				msg:        "failed to update item: not found",
			},
			{
				query:  postgresUpdateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
				args:       []driver.Value{uuid.NullUUID{}, locationPlayerID, uuid.NullUUID{}},
				err:        &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to update item: bad request: the given ownerID or locationID does not exist: "+
					"ownerID '%s', locationID '%s (%s)'", ownerID, locationPlayerID, "player"),
			},
			{
				query:  postgresUpdateQ,
				driver: postgres.ItemDriver{},
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

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.ItemUpdate{
				ItemChange: asset.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
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
			locationID asset.ItemLocationID
			args       []driver.Value
		}{
			{
				query:  postgresUpdateQ,
				driver: postgres.ItemDriver{},
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

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.ItemUpdate{
				ItemChange: asset.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
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
			locationID asset.ItemLocationID
			args       []driver.Value
		}{
			{
				query:  postgresUpdateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationItemID, nil, nil, created, updated),
				locationID: locationItemID,
				args:       []driver.Value{locationItemID, uuid.NullUUID{}, uuid.NullUUID{}},
			},
			{
				query:  postgresUpdateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, locationPlayerID, nil, created, updated),
				locationID: locationPlayerID,
				args:       []driver.Value{uuid.NullUUID{}, locationPlayerID, uuid.NullUUID{}},
			},
			{
				query:  postgresUpdateQ,
				driver: postgres.ItemDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_item_id", "location_player_id", "location_room_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, nil, nil, locationRoomID, created, updated),
				locationID: locationRoomID,
				args:       []driver.Value{uuid.NullUUID{}, uuid.NullUUID{}, locationRoomID},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.ItemUpdate{
				ItemChange: asset.ItemChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: test.locationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, test.args[0], test.args[1], test.args[2]).WillReturnRows(test.rows)

			item, err := i.Update(ctx, id, update)

			assert.Nil(t, err)
			assert.Compare(t, *item, asset.Item{
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
		postgresRemoveQ = `^DELETE FROM items WHERE id = (.+)$`
	)

	var (
		ctx = context.Background()
		id  = asset.ItemID(uuid.New())
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.ItemDriver
			err    error
			msg    string
		}{
			{
				query:  postgresRemoveQ,
				driver: postgres.ItemDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to remove item: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
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
				query:  postgresRemoveQ,
				driver: postgres.ItemDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			i := data.ItemStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

			err = i.Remove(ctx, id)

			assert.Nil(t, err)
			assert.MockExpectationsMet(t, mock)
		}
	})
}
