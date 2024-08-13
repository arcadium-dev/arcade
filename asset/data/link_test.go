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

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/data"
	"arcadium.dev/arcade/asset/data/postgres"
)

func TestLinksList(t *testing.T) {
	const (
		postgresListQ              = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links$"
		postgresListWithAllFilterQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links" +
			" WHERE owner_id = (.+) AND location_id (.+) AND destination_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithOwnerFilterQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links" +
			" WHERE owner_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithLocationFilterQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links" +
			" WHERE location_id = (.+) LIMIT (.+) OFFSET (.+)$"
		postgresListWithDestinationFilterQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links" +
			" WHERE destination_id = (.+) LIMIT (.+) OFFSET (.+)$"
	)

	var (
		ctx           = context.Background()
		id            = asset.LinkID(uuid.New())
		name          = "Going Nowhere"
		desc          = "Like your life."
		ownerID       = asset.PlayerID(uuid.New())
		locationID    = asset.RoomID(uuid.New())
		destinationID = asset.RoomID(uuid.New())
		created       = arcade.Timestamp{Time: time.Now()}
		updated       = arcade.Timestamp{Time: time.Now()}
	)

	t.Run("query error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
		}{
			{
				query:  postgresListQ,
				driver: postgres.LinkDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnError(errors.New("query error"))

			_, err = l.List(ctx, asset.LinkFilter{})

			assert.Error(t, err, `failed to list links: internal server error: query error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("sql scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  postgresListQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			_, err = l.List(context.Background(), asset.LinkFilter{})

			assert.Error(t, err, `failed to list links: internal server error: scan error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			filter asset.LinkFilter
			driver data.LinkDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  postgresListQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
			{
				query: postgresListWithAllFilterQ,
				filter: asset.LinkFilter{
					OwnerID:       ownerID,
					LocationID:    locationID,
					DestinationID: destinationID,
					Limit:         asset.DefaultLinkFilterLimit,
					Offset:        10,
				},
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
			{
				query: postgresListWithOwnerFilterQ,
				filter: asset.LinkFilter{
					OwnerID: ownerID,
					Limit:   asset.DefaultLinkFilterLimit,
					Offset:  10,
				},
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
			{
				query: postgresListWithLocationFilterQ,
				filter: asset.LinkFilter{
					LocationID: locationID,
					Limit:      asset.DefaultLinkFilterLimit,
					Offset:     10,
				},
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
			{
				query: postgresListWithDestinationFilterQ,
				filter: asset.LinkFilter{
					DestinationID: destinationID,
					Limit:         asset.DefaultLinkFilterLimit,
					Offset:        10,
				},
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows).RowsWillBeClosed()

			links, err := l.List(context.Background(), test.filter)

			assert.Nil(t, err)
			assert.Equal(t, len(links), 1)
			assert.Compare(t, *links[0], asset.Link{
				ID:            id,
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
				Created:       created,
				Updated:       updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestLinksGet(t *testing.T) {
	const (
		postgresGetQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links WHERE id = (.+)$"
	)

	var (
		ctx           = context.Background()
		id            = asset.LinkID(uuid.New())
		name          = "Going Nowhere"
		desc          = "Like your life."
		ownerID       = asset.PlayerID(uuid.New())
		locationID    = asset.RoomID(uuid.New())
		destinationID = asset.RoomID(uuid.New())
		created       = arcade.Timestamp{Time: time.Now()}
		updated       = arcade.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			err    error
			msg    string
		}{
			{
				query:  postgresGetQ,
				driver: postgres.LinkDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to get link: not found",
			},
			{
				query:  postgresGetQ,
				driver: postgres.LinkDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to get link: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WithArgs(id).WillReturnError(test.err)

			_, err = l.Get(ctx, id)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  postgresGetQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnRows(test.rows)

			link, err := l.Get(ctx, id)

			assert.Nil(t, err)
			assert.Compare(t, *link, asset.Link{
				ID:            id,
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
				Created:       created,
				Updated:       updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestLinksCreate(t *testing.T) {
	const (
		postgresCreateQ = `^INSERT INTO links \(name, description, owner_id, location_id, destination_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING id, name, description, owner_id, location_id, destination_id, created, updated$`
	)

	var (
		ctx           = context.Background()
		id            = asset.LinkID(uuid.New())
		name          = "Going Nowhere"
		desc          = "Like your life."
		ownerID       = asset.PlayerID(uuid.New())
		locationID    = asset.RoomID(uuid.New())
		destinationID = asset.RoomID(uuid.New())
		created       = arcade.Timestamp{Time: time.Now()}
		updated       = arcade.Timestamp{Time: time.Now()}
	)

	t.Run("violations", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			rows   *sqlmock.Rows
			err    error
			msg    string
		}{
			{
				query:  postgresCreateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to create link: bad request: the given ownerID, locationID or destinationID does not exist: "+
					"ownerID '%s', locationID '%s', destinationID '%s'", ownerID, locationID, destinationID),
			},
			{
				query:  postgresCreateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg: "failed to create link: bad request: link name 'Going Nowhere' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := asset.LinkCreate{
				LinkChange: asset.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, locationID, destinationID).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = l.Create(ctx, create)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  postgresCreateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := asset.LinkCreate{
				LinkChange: asset.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, locationID, destinationID).WillReturnRows(test.rows)

			_, err = l.Create(ctx, create)

			assert.Error(t, err, "failed to create link: internal server error: scan error")
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  postgresCreateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}

			create := asset.LinkCreate{
				LinkChange: asset.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, locationID, destinationID).WillReturnRows(test.rows)

			link, err := l.Create(ctx, create)

			assert.Nil(t, err)
			assert.Compare(t, *link, asset.Link{
				ID:            id,
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
				Created:       created,
				Updated:       updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestLinksUpdate(t *testing.T) {
	const (
		postgresUpdateQ = `^UPDATE links SET name = (.+), description = (.+), owner_id = (.+), location_id = (.+), destination_id = (.+) ` +
			`WHERE id = (.+) ` +
			`RETURNING id, name, description, owner_id, location_id, destination_id, created, updated$`
	)

	var (
		ctx           = context.Background()
		id            = asset.LinkID(uuid.New())
		name          = "Going Nowhere"
		desc          = "Like your life."
		ownerID       = asset.PlayerID(uuid.New())
		locationID    = asset.RoomID(uuid.New())
		destinationID = asset.RoomID(uuid.New())
		created       = arcade.Timestamp{Time: time.Now()}
		updated       = arcade.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			rows   *sqlmock.Rows
			err    error
			msg    string
		}{
			{
				query:  postgresUpdateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
				err: sql.ErrNoRows,
				msg: "failed to update link: not found",
			},
			{
				query:  postgresUpdateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to update link: bad request: the given ownerID, locationID or destinationID does not exist: "+
					"ownerID '%s', locationID '%s', destinationID '%s'", ownerID, locationID, destinationID),
			},
			{
				query:  postgresUpdateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.UniqueViolation},
				msg: "failed to update link: bad request: link name 'Going Nowhere' already exists",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.LinkUpdate{
				LinkChange: asset.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, locationID, destinationID).WillReturnRows(test.rows).WillReturnError(test.err)

			_, err = l.Update(ctx, id, update)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  postgresUpdateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.LinkUpdate{
				LinkChange: asset.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, locationID, destinationID).WillReturnRows(test.rows)

			_, err = l.Update(ctx, id, update)

			assert.Error(t, err, "failed to update link: internal server error: scan error")
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  postgresUpdateQ,
				driver: postgres.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := asset.LinkUpdate{
				LinkChange: asset.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, locationID, destinationID).WillReturnRows(test.rows)

			link, err := l.Update(ctx, id, update)

			assert.Nil(t, err)
			assert.Compare(t, *link, asset.Link{
				ID:            id,
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
				Created:       created,
				Updated:       updated,
			}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			assert.MockExpectationsMet(t, mock)
		}
	})
}

func TestLinksRemove(t *testing.T) {
	const (
		postgresRemoveQ = `^DELETE FROM links WHERE id = (.+)$`
	)

	var (
		ctx = context.Background()
		id  = asset.LinkID(uuid.New())
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			err    error
			msg    string
		}{
			{
				query:  postgresRemoveQ,
				driver: postgres.LinkDriver{},
				err:    errors.New("unknown error"),
				msg:    "failed to remove link: internal server error: unknown error",
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnError(test.err)

			err = l.Remove(ctx, id)

			assert.Error(t, err, test.msg)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
		}{
			{
				query:  postgresRemoveQ,
				driver: postgres.LinkDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectExec(test.query).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

			err = l.Remove(ctx, id)

			assert.Nil(t, err)
			assert.MockExpectationsMet(t, mock)
		}
	})
}
