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
	"arcadium.dev/core/require"
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/data"
	"arcadium.dev/arcade/data/postgres"
	"arcadium.dev/arcade/user"
)

func TestUserStorageList(t *testing.T) {
	t.Parallel()
	const (
		postgresListQ           = "^SELECT id, login, public_key, player_id, created, updated FROM users$"
		postgresListWithFilterQ = "^SELECT id, login, public_key, player_id, created, updated FROM users LIMIT (.+) OFFSET (.+)$"
	)

	var (
		id        = user.ID(uuid.New())
		login     = "ajones"
		publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
		playerID  = asset.PlayerID(uuid.New())
		created   = arcade.Timestamp{Time: time.Now()}
		updated   = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name      string
		query     string
		filter    user.Filter
		mockSetup func(sqlmock.Sqlmock, string)
		verify    func(*testing.T, sqlmock.Sqlmock, []*user.User, error)
	}{
		{
			name:  "query failure",
			query: postgresListQ,
			mockSetup: func(mock sqlmock.Sqlmock, query string) {
				mock.ExpectQuery(query).WillReturnError(errors.New("query failure"))
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, users []*user.User, err error) {
				assert.Nil(t, users)
				assert.Error(t, err, "failed to list users: internal server error: query failure")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name:  "rows failure",
			query: postgresListQ,
			mockSetup: func(mock sqlmock.Sqlmock, query string) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_kd", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated).
					RowError(0, errors.New("rows failure"))
				mock.ExpectQuery(query).WillReturnRows(rows).RowsWillBeClosed()
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, users []*user.User, err error) {
				assert.Nil(t, users)
				assert.Error(t, err, "failed to list users: internal server error: rows failure")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name:  "success - without filter",
			query: postgresListQ,
			mockSetup: func(mock sqlmock.Sqlmock, query string) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_kd", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated)
				mock.ExpectQuery(query).WillReturnRows(rows).RowsWillBeClosed()
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, users []*user.User, err error) {
				assert.Nil(t, err)
				require.Equal(t, len(users), 1)
				assert.Compare(t, *users[0], user.User{
					ID:        id,
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "success - with filter",
			filter: user.Filter{
				Offset: 1,
				Limit:  1,
			},
			query: postgresListWithFilterQ,
			mockSetup: func(mock sqlmock.Sqlmock, query string) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_kd", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated)
				mock.ExpectQuery(query).WillReturnRows(rows).RowsWillBeClosed()
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, users []*user.User, err error) {
				require.Nil(t, err)
				assert.Equal(t, len(users), 1)
				assert.Compare(t, *users[0], user.User{
					ID:        id,
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
				assert.MockExpectationsMet(t, mock)
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			test.mockSetup(mock, test.query)

			us := data.UserStorage{
				DB: &sql.DB{DB: db},
				Driver: postgres.UserDriver{
					Driver: postgres.Driver{},
				},
			}

			users, err := us.List(context.Background(), test.filter)
			test.verify(t, mock, users, err)
		})
	}
}

func TestUserStorageGet(t *testing.T) {
	t.Parallel()
	const (
		postgresGetQ = "^SELECT id, login, public_key, player_id, created, updated FROM users WHERE id = (.+)$"
	)

	var (
		id        = user.ID(uuid.New())
		login     = "ajones"
		publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
		playerID  = asset.PlayerID(uuid.New())
		created   = arcade.Timestamp{Time: time.Now()}
		updated   = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name      string
		mockSetup func(sqlmock.Sqlmock)
		verify    func(*testing.T, sqlmock.Sqlmock, *user.User, error)
	}{
		{
			name: "user not found failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(postgresGetQ).WithArgs(id).WillReturnError(sql.ErrNoRows)
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, "failed to get user: not found")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "query failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(postgresGetQ).WithArgs(id).WillReturnError(errors.New("query failure"))
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, "failed to get user: internal server error: query failure")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "success",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_id", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated)
				mock.ExpectQuery(postgresGetQ).WillReturnRows(rows)
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				require.Nil(t, err)
				require.NotNil(t, u)
				assert.Compare(t, *u, user.User{
					ID:        id,
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
				assert.MockExpectationsMet(t, mock)
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			test.mockSetup(mock)

			us := data.UserStorage{
				DB: &sql.DB{DB: db},
				Driver: postgres.UserDriver{
					Driver: postgres.Driver{},
				},
			}

			u, err := us.Get(context.Background(), id)
			test.verify(t, mock, u, err)
		})
	}
}

func TestUserStorageCreate(t *testing.T) {
	t.Parallel()
	const (
		postgresCreateQ = `^INSERT INTO users \(login, public_key, player_id\) ` +
			`VALUES \((.+), (.+), (.+)\) ` +
			`RETURNING id, login, public_key, player_id, created, updated$`
	)

	var (
		id        = user.ID(uuid.New())
		login     = "ajones"
		publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
		playerID  = asset.PlayerID(uuid.New())
		created   = arcade.Timestamp{Time: time.Now()}
		updated   = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name      string
		mockSetup func(sqlmock.Sqlmock)
		verify    func(*testing.T, sqlmock.Sqlmock, *user.User, error)
	}{
		{
			name: "foreign key violation - player does not exist",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_id", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated)
				mock.ExpectQuery(postgresCreateQ).
					WithArgs(login, publicKey, playerID).
					WillReturnRows(rows).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, fmt.Sprintf("failed to create user: bad request: the given playerID does not exist, playerID: '%s'", playerID))
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "unique violation - player already exists",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_id", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated)
				mock.ExpectQuery(postgresCreateQ).
					WithArgs(login, publicKey, playerID).
					WillReturnRows(rows).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, "failed to create user: bad request: user login 'ajones' already exists")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "scan failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_id", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated).
					RowError(0, errors.New("scan failure"))
				mock.ExpectQuery(postgresCreateQ).
					WithArgs(login, publicKey, playerID).
					WillReturnRows(rows)
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, "failed to create user: internal server error: scan failure")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "success",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_id", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated)
				mock.ExpectQuery(postgresCreateQ).
					WithArgs(login, publicKey, playerID).
					WillReturnRows(rows)
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				require.Nil(t, err)
				require.NotNil(t, u)
				assert.Compare(t, *u, user.User{
					ID:        id,
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
				assert.MockExpectationsMet(t, mock)
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			test.mockSetup(mock)

			create := user.Create{Change: user.Change{Login: login, PublicKey: publicKey, PlayerID: playerID}}

			us := data.UserStorage{
				DB: &sql.DB{DB: db},
				Driver: postgres.UserDriver{
					Driver: postgres.Driver{},
				},
			}

			u, err := us.Create(context.Background(), create)
			test.verify(t, mock, u, err)
		})
	}
}

func TestUserStorageUpdate(t *testing.T) {
	t.Parallel()
	const (
		postgresUpdateQ = `^UPDATE users SET login = (.+), public_key = (.+), player_id = (.+) ` +
			`WHERE id = (.+) ` +
			`RETURNING id, login, public_key, player_id, created, updated$`
	)

	var (
		id        = user.ID(uuid.New())
		login     = "ajones"
		publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
		playerID  = asset.PlayerID(uuid.New())
		created   = arcade.Timestamp{Time: time.Now()}
		updated   = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name      string
		mockSetup func(sqlmock.Sqlmock)
		verify    func(*testing.T, sqlmock.Sqlmock, *user.User, error)
	}{
		{
			name: "user not found failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(postgresUpdateQ).WithArgs(id, login, publicKey, playerID).WillReturnError(sql.ErrNoRows)
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, "failed to update user: not found")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "foreign key violation - player does not exist",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(postgresUpdateQ).WithArgs(id, login, publicKey, playerID).WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, fmt.Sprintf("failed to update user: bad request: the given playerID not exist, playerID: '%s'", playerID))
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "unique violation - player already exists",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(postgresUpdateQ).WithArgs(id, login, publicKey, playerID).WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, "failed to update user: bad request: user login 'ajones' already exists")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "scan failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_id", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated).
					RowError(0, errors.New("scan failure"))
				mock.ExpectQuery(postgresUpdateQ).
					WithArgs(id, login, publicKey, playerID).
					WillReturnRows(rows)
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, u)
				assert.Error(t, err, "failed to update user: internal server error: scan failure")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "success",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "login", "public_key", "player_id", "created", "updated"}).
					AddRow(id, login, publicKey, playerID, created, updated)
				mock.ExpectQuery(postgresUpdateQ).
					WithArgs(id, login, publicKey, playerID).
					WillReturnRows(rows)
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, u *user.User, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, u)
				assert.Compare(t, *u, user.User{
					ID:        id,
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
				assert.MockExpectationsMet(t, mock)
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			test.mockSetup(mock)

			update := user.Update{Change: user.Change{Login: login, PublicKey: publicKey, PlayerID: playerID}}

			us := data.UserStorage{
				DB: &sql.DB{DB: db},
				Driver: postgres.UserDriver{
					Driver: postgres.Driver{},
				},
			}

			u, err := us.Update(context.Background(), id, update)
			test.verify(t, mock, u, err)
		})
	}
}

func TestUsersRemove(t *testing.T) {
	t.Parallel()
	const (
		postgresRemoveQ = `^DELETE FROM users WHERE id = (.+)$`
	)

	var (
		id = user.ID(uuid.New())
	)

	tests := []struct {
		name      string
		mockSetup func(sqlmock.Sqlmock)
		verify    func(*testing.T, sqlmock.Sqlmock, error)
	}{
		{
			name: "query failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(postgresRemoveQ).WithArgs(id).WillReturnError(errors.New("query failure"))
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, err error) {
				assert.Error(t, err, "failed to remove user: internal server error: query failure")
				assert.MockExpectationsMet(t, mock)
			},
		},
		{
			name: "success",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(postgresRemoveQ).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, err error) {
				assert.Nil(t, err)
				assert.MockExpectationsMet(t, mock)
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New()
			assert.Nil(t, err)

			test.mockSetup(mock)

			us := data.UserStorage{
				DB: &sql.DB{DB: db},
				Driver: postgres.UserDriver{
					Driver: postgres.Driver{},
				},
			}

			err = us.Remove(context.Background(), id)
			test.verify(t, mock, err)
		})
	}
}
