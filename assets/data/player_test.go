package persistence_test

/*
import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/storage"
	"arcadium.dev/arcade/storage/cockroach"
)

func TestPlayersList(t *testing.T) {
	const (
		listQ = "^SELECT player_id, name, description, home_id, location_id, created, updated FROM players$"
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		homeID      = uuid.NewString()
		locationID  = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("sql query error", func(t *testing.T) {
		p, mock := setupPlayers(t)
		mock.ExpectQuery(listQ).
			WillReturnError(errors.New("unknown error"))

		_, err := p.List(context.Background(), arcade.PlayersFilter{})

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to list players: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("sql scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"player_id", "name", "description", "home_id", "location_id", "created", "updated",
		}).
			AddRow(id, name, description, homeID, locationID, created, updated).
			RowError(0, errors.New("scan error"))

		p, mock := setupPlayers(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		_, err := p.List(context.Background(), arcade.PlayersFilter{})

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to list players: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated)

		p, mock := setupPlayers(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		players, err := p.List(context.Background(), arcade.PlayersFilter{})

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(players) != 1 {
			t.Fatalf("Unexpected length of player list")
		}
		if players[0].ID != id ||
			players[0].Name != name ||
			players[0].Description != description ||
			players[0].HomeID != homeID ||
			players[0].LocationID != locationID {
			t.Errorf("\nExpected player: %+v", players[0])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestPlayersGet(t *testing.T) {
	const (
		getQ = "^SELECT player_id, name, description, home_id, location_id, created, updated FROM players WHERE player_id = (.+)$"
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		homeID      = uuid.NewString()
		locationID  = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("invalid playerID", func(t *testing.T) {
		p, _ := setupPlayers(t)

		_, err := p.Get(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get player: invalid argument: invalid player id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		p, mock := setupPlayers(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(sql.ErrNoRows)

		_, err := p.Get(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get player: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		p, mock := setupPlayers(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(errors.New("unknown error"))

		_, err := p.Get(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get player: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated)

		p, mock := setupPlayers(t)
		mock.ExpectQuery(getQ).WillReturnRows(rows)

		player, err := p.Get(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if player.ID != id ||
			player.Name != name ||
			player.Description != description ||
			player.HomeID != homeID ||
			player.LocationID != locationID {
			t.Errorf("\nExpected player: %+v", player)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestPlayersCreate(t *testing.T) {
	const (
		createQ = `^INSERT INTO players \(name, description, home_id, location_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING player_id, name, description, home_id, location_id, created, updated$`
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		homeID      = "00000000-0000-0000-0000-000000000001"
		locationID  = "00000000-0000-0000-0000-000000000001"
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("empty name", func(t *testing.T) {
		req := arcade.PlayerRequest{Description: description, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: empty player name"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long name", func(t *testing.T) {
		n := ""
		for i := 0; i <= arcade.MaxPlayerNameLen; i++ {
			n += "a"
		}
		req := arcade.PlayerRequest{Name: n, Description: description, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: player name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: empty player description"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long description", func(t *testing.T) {
		d := ""
		for i := 0; i <= arcade.MaxPlayerDescriptionLen; i++ {
			d += "a"
		}
		req := arcade.PlayerRequest{Name: name, Description: d, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: player description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid homeID", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: "42", LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: invalid homeID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid location", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: "42"}

		p, _ := setupPlayers(t)

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: invalid locationID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated)

		p, mock := setupPlayers(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, homeID, locationID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: the given homeID or locationID does not exist: " +
			"homeID '00000000-0000-0000-0000-000000000001', locationID '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated)

		p, mock := setupPlayers(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, homeID, locationID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: already exists: player already exists"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated).
			RowError(0, errors.New("scan error"))

		p, mock := setupPlayers(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, homeID, locationID).
			WillReturnRows(row)

		_, err := p.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated)

		p, mock := setupPlayers(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, homeID, locationID).
			WillReturnRows(row)

		player, err := p.Create(context.Background(), req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if player.ID != id ||
			player.Name != name ||
			player.Description != description ||
			player.HomeID != homeID ||
			player.LocationID != locationID {
			t.Errorf("\nExpected player: %+v", player)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestPlayersUpdate(t *testing.T) {
	const (
		// updateQ = `^UPDATE players SET (.+) WHERE (.+) RETURNING (.+)$`
		updateQ = `^UPDATE players SET name = (.+), description = (.+), home_id = (.+), location_id = (.+) ` +
			`WHERE player_id = (.+) ` +
			`RETURNING player_id, name, description, home_id, location_id, created, updated$`
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		homeID      = "00000000-0000-0000-0000-000000000001"
		locationID  = "00000000-0000-0000-0000-000000000001"
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("invalid player id", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Update(context.Background(), "42", req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: invalid player id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := arcade.PlayerRequest{Description: description, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: empty player name"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long name", func(t *testing.T) {
		n := ""
		for i := 0; i <= arcade.MaxPlayerNameLen; i++ {
			n += "a"
		}
		req := arcade.PlayerRequest{Name: n, Description: description, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: player name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: empty player description"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long description", func(t *testing.T) {
		d := ""
		for i := 0; i <= arcade.MaxPlayerDescriptionLen; i++ {
			d += "a"
		}
		req := arcade.PlayerRequest{Name: name, Description: d, HomeID: homeID, LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: player description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid home", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: "42", LocationID: locationID}

		p, _ := setupPlayers(t)

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: invalid homeID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid location", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: "42"}

		p, _ := setupPlayers(t)

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: invalid locationID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}

		p, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, homeID, locationID).
			WillReturnError(sql.ErrNoRows)

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated)

		p, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, homeID, locationID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: the given homeID or locationID does not exist: " +
			"homeID '00000000-0000-0000-0000-000000000001', locationID '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated)

		p, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, homeID, locationID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: already exists: player name is not unique"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated).
			RowError(0, errors.New("scan error"))

		p, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, homeID, locationID).
			WillReturnRows(row)

		_, err := p.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := arcade.PlayerRequest{Name: name, Description: description, HomeID: homeID, LocationID: locationID}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home_id", "location_id", "created", "updated"}).
			AddRow(id, name, description, homeID, locationID, created, updated)

		p, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, homeID, locationID).
			WillReturnRows(row)

		player, err := p.Update(context.Background(), id, req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if player.ID != id ||
			player.Name != name ||
			player.Description != description ||
			player.HomeID != homeID ||
			player.LocationID != locationID {
			t.Errorf("\nExpected player: %+v", player)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestPlayersRemove(t *testing.T) {
	const (
		removeQ = `^DELETE FROM players WHERE player_id = (.+)$`
	)

	var (
		id = uuid.NewString()
	)

	t.Run("invalid player id", func(t *testing.T) {
		p, _ := setupPlayers(t)

		err := p.Remove(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove player: invalid argument: invalid player id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		p, mock := setupPlayers(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		err := p.Remove(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove player: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		p, mock := setupPlayers(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(errors.New("unknown error"))

		err := p.Remove(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove player: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		p, mock := setupPlayers(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := p.Remove(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected err: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func setupPlayers(t *testing.T) (storage.Players, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create sqlmock db")
	}

	return storage.Players{DB: db, Driver: cockroach.Driver{}}, mock
}
*/
