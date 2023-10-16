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

	"arcadium.dev/arcade/assets"
	"arcadium.dev/arcade/assets/data"
	"arcadium.dev/arcade/assets/data/cockroach"
)

func TestLinksList(t *testing.T) {
	const (
		cockroachListQ              = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links$"
		cockroachListWithAllFilterQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links" +
			" WHERE owner_id = (.+) AND location_id (.+) AND destination_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithOwnerFilterQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links" +
			" WHERE owner_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithLocationFilterQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links" +
			" WHERE location_id = (.+) LIMIT (.+) OFFSET (.+)$"
		cockroachListWithDestinationFilterQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links" +
			" WHERE destination_id = (.+) LIMIT (.+) OFFSET (.+)$"
	)

	var (
		ctx           = context.Background()
		id            = assets.LinkID(uuid.New())
		name          = "Going Nowhere"
		desc          = "Like your life."
		ownerID       = assets.PlayerID(uuid.New())
		locationID    = assets.RoomID(uuid.New())
		destinationID = assets.RoomID(uuid.New())
		created       = assets.Timestamp{Time: time.Now()}
		updated       = assets.Timestamp{Time: time.Now()}
	)

	t.Run("query error", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.LinkDriver{},
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			mock.ExpectQuery(test.query).WillReturnError(errors.New("query error"))

			_, err = l.List(ctx, assets.LinkFilter{})

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
				query:  cockroachListQ,
				driver: cockroach.LinkDriver{},
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

			_, err = l.List(context.Background(), assets.LinkFilter{})

			assert.Error(t, err, `failed to list links: internal server error: scan error`)
			assert.MockExpectationsMet(t, mock)
		}
	})

	t.Run("success", func(t *testing.T) {
		tests := []struct {
			query  string
			filter assets.LinkFilter
			driver data.LinkDriver
			rows   *sqlmock.Rows
		}{
			{
				query:  cockroachListQ,
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
			{
				query: cockroachListWithAllFilterQ,
				filter: assets.LinkFilter{
					OwnerID:       ownerID,
					LocationID:    locationID,
					DestinationID: destinationID,
					Limit:         assets.DefaultLinkFilterLimit,
					Offset:        10,
				},
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
			{
				query: cockroachListWithOwnerFilterQ,
				filter: assets.LinkFilter{
					OwnerID: ownerID,
					Limit:   assets.DefaultLinkFilterLimit,
					Offset:  10,
				},
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
			{
				query: cockroachListWithLocationFilterQ,
				filter: assets.LinkFilter{
					LocationID: locationID,
					Limit:      assets.DefaultLinkFilterLimit,
					Offset:     10,
				},
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
			{
				query: cockroachListWithDestinationFilterQ,
				filter: assets.LinkFilter{
					DestinationID: destinationID,
					Limit:         assets.DefaultLinkFilterLimit,
					Offset:        10,
				},
				driver: cockroach.LinkDriver{},
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
			assert.Compare(t, *links[0], assets.Link{
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
		cockroachGetQ = "^SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links WHERE id = (.+)$"
	)

	var (
		ctx           = context.Background()
		id            = assets.LinkID(uuid.New())
		name          = "Going Nowhere"
		desc          = "Like your life."
		ownerID       = assets.PlayerID(uuid.New())
		locationID    = assets.RoomID(uuid.New())
		destinationID = assets.RoomID(uuid.New())
		created       = assets.Timestamp{Time: time.Now()}
		updated       = assets.Timestamp{Time: time.Now()}
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			err    error
			msg    string
		}{
			{
				query:  cockroachGetQ,
				driver: cockroach.LinkDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to get link: not found",
			},
			{
				query:  cockroachGetQ,
				driver: cockroach.LinkDriver{},
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
				query:  cockroachGetQ,
				driver: cockroach.LinkDriver{},
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
			assert.Compare(t, *link, assets.Link{
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
		cockroachCreateQ = `^INSERT INTO links \(name, description, owner_id, location_id, destination_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING id, name, description, owner_id, location_id, destination_id, created, updated$`
	)

	var (
		ctx           = context.Background()
		id            = assets.LinkID(uuid.New())
		name          = "Going Nowhere"
		desc          = "Like your life."
		ownerID       = assets.PlayerID(uuid.New())
		locationID    = assets.RoomID(uuid.New())
		destinationID = assets.RoomID(uuid.New())
		created       = assets.Timestamp{Time: time.Now()}
		updated       = assets.Timestamp{Time: time.Now()}
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
				query:  cockroachCreateQ,
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to create link: bad request: the given ownerID, locationID or destinationID does not exist: "+
					"ownerID '%s', locationID '%s', destinationID '%s'", ownerID, locationID, destinationID),
			},
			{
				query:  cockroachCreateQ,
				driver: cockroach.LinkDriver{},
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
			create := assets.LinkCreate{
				LinkChange: assets.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
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
				query:  cockroachCreateQ,
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			create := assets.LinkCreate{
				LinkChange: assets.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
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
				query:  cockroachCreateQ,
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}

			create := assets.LinkCreate{
				LinkChange: assets.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
			}
			mock.ExpectQuery(test.query).WithArgs(name, desc, ownerID, locationID, destinationID).WillReturnRows(test.rows)

			link, err := l.Create(ctx, create)

			assert.Nil(t, err)
			assert.Compare(t, *link, assets.Link{
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
		cockroachUpdateQ = `^UPDATE links SET name = (.+), description = (.+), owner_id = (.+), location_id = (.+), destination_id = (.+) ` +
			`WHERE id = (.+) ` +
			`RETURNING id, name, description, owner_id, location_id, destination_id, created, updated$`
	)

	var (
		ctx           = context.Background()
		id            = assets.LinkID(uuid.New())
		name          = "Going Nowhere"
		desc          = "Like your life."
		ownerID       = assets.PlayerID(uuid.New())
		locationID    = assets.RoomID(uuid.New())
		destinationID = assets.RoomID(uuid.New())
		created       = assets.Timestamp{Time: time.Now()}
		updated       = assets.Timestamp{Time: time.Now()}
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
				query:  cockroachUpdateQ,
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
				err: sql.ErrNoRows,
				msg: "failed to update link: not found",
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
				err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
				msg: fmt.Sprintf("failed to update link: bad request: the given ownerID, locationID or destinationID does not exist: "+
					"ownerID '%s', locationID '%s', destinationID '%s'", ownerID, locationID, destinationID),
			},
			{
				query:  cockroachUpdateQ,
				driver: cockroach.LinkDriver{},
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
			update := assets.LinkUpdate{
				LinkChange: assets.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
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
				query:  cockroachUpdateQ,
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated).
					RowError(0, errors.New("scan error")),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := assets.LinkUpdate{
				LinkChange: assets.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
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
				query:  cockroachUpdateQ,
				driver: cockroach.LinkDriver{},
				rows: sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
					AddRow(id, name, desc, ownerID, locationID, destinationID, created, updated),
			},
		}

		for _, test := range tests {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			l := data.LinkStorage{DB: &sql.DB{DB: db}, Driver: test.driver}
			update := assets.LinkUpdate{
				LinkChange: assets.LinkChange{Name: name, Description: desc, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID},
			}
			mock.ExpectQuery(test.query).WithArgs(id, name, desc, ownerID, locationID, destinationID).WillReturnRows(test.rows)

			link, err := l.Update(ctx, id, update)

			assert.Nil(t, err)
			assert.Compare(t, *link, assets.Link{
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
		cockroachRemoveQ = `^DELETE FROM links WHERE id = (.+)$`
	)

	var (
		ctx = context.Background()
		id  = assets.LinkID(uuid.New())
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			query  string
			driver data.LinkDriver
			err    error
			msg    string
		}{
			{
				query:  cockroachRemoveQ,
				driver: cockroach.LinkDriver{},
				err:    sql.ErrNoRows,
				msg:    "failed to remove link: not found",
			},
			{
				query:  cockroachRemoveQ,
				driver: cockroach.LinkDriver{},
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
				query:  cockroachRemoveQ,
				driver: cockroach.LinkDriver{},
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
