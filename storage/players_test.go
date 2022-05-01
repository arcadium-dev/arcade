//  Copyright 2022 arcadium.dev <info@arcadium.dev>
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

func TestPlayersList(t *testing.T) {
	const (
		listQ = "^SELECT player_id, name, description, home, location, created, updated FROM players$"
	)

	t.Run("sql query error", func(t *testing.T) {
		s, mock := setupPlayers(t)
		mock.ExpectQuery(listQ).
			WillReturnError(errors.New("unknown error"))

		_, err := s.List(context.Background())

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
			"player_id", "name", "description", "home", "location", "created", "updated",
		}).
			AddRow(id, name, description, home, location, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupPlayers(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		_, err := s.List(context.Background())

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
		rows := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupPlayers(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		players, err := s.list(context.Background())

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(players) != 1 {
			t.Fatalf("Unexpected length of player list")
		}
		if players[0].ID() != id ||
			players[0].Name() != name ||
			players[0].Description() != description ||
			players[0].Home() != home ||
			players[0].Location() != location {
			t.Errorf("\nExpected player: %+v", players[0])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestPlayersGet(t *testing.T) {
	const (
		getQ = "^SELECT player_id, name, description, home, location, created, updated FROM players WHERE player_id = (.+)$"
	)

	t.Run("invalid playerID", func(t *testing.T) {
		s, _ := setupPlayers(t)

		_, err := s.get(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get player: invalid argument: invalid player id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, mock := setupPlayers(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(sql.ErrNoRows)

		_, err := s.get(context.Background(), id)

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
		s, mock := setupPlayers(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(errors.New("unknown error"))

		_, err := s.get(context.Background(), id)

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
		rows := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupPlayers(t)
		mock.ExpectQuery(getQ).WillReturnRows(rows)

		p, err := s.get(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if p.ID() != id ||
			p.Name() != name ||
			p.Description() != description ||
			p.Home() != home ||
			p.Location() != location {
			t.Errorf("\nExpected player: %+v", p)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestPlayersCreate(t *testing.T) {
	const (
		createQ = `^INSERT INTO players \(name, description, home, location\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING player_id, name, description, home, location, created, updated$`
	)

	t.Run("empty name", func(t *testing.T) {
		req := playerRequest{Description: description, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.create(context.Background(), req)

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
		for i := 0; i <= maxNameLen; i++ {
			n += "a"
		}
		req := playerRequest{Name: n, Description: description, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: player name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := playerRequest{Name: name, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.create(context.Background(), req)

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
		for i := 0; i <= maxDescriptionLen; i++ {
			d += "a"
		}
		req := playerRequest{Name: name, Description: d, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: player description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid home", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: "42", Location: location}

		s, _ := setupPlayers(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: invalid home: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid location", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: home, Location: "42"}

		s, _ := setupPlayers(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: invalid location: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupPlayers(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, home, location).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create player: invalid argument: the given home or location does not exist: " +
			"home '00000000-0000-0000-0000-000000000001', location '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupPlayers(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, home, location).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := s.create(context.Background(), req)

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
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupPlayers(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, home, location).
			WillReturnRows(row)

		_, err := s.create(context.Background(), req)

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
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupPlayers(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, home, location).
			WillReturnRows(row)

		_, err := s.create(context.Background(), req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if p.ID() != id ||
			p.Name() != name ||
			p.Description() != description ||
			p.Home() != home ||
			p.Location() != location {
			t.Errorf("\nExpected player: %+v", p)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestPlayersUpdate(t *testing.T) {
	const (
		// updateQ = `^UPDATE players SET (.+) WHERE (.+) RETURNING (.+)$`
		updateQ = `^UPDATE players SET name = (.+), description = (.+), home = (.+), location = (.+) ` +
			`WHERE player_id = (.+) ` +
			`RETURNING player_id, name, description, home, location, created, updated$`
	)

	t.Run("invalid player id", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.update(context.Background(), "42", req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: invalid player id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := playerRequest{Description: description, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.update(context.Background(), id, req)

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
		for i := 0; i <= maxNameLen; i++ {
			n += "a"
		}
		req := playerRequest{Name: n, Description: description, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: player name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := playerRequest{Name: name, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.update(context.Background(), id, req)

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
		for i := 0; i <= maxDescriptionLen; i++ {
			d += "a"
		}
		req := playerRequest{Name: name, Description: d, Home: home, Location: location}

		s, _ := setupPlayers(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: player description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid home", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: "42", Location: location}

		s, _ := setupPlayers(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: invalid home: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid location", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: home, Location: "42"}

		s, _ := setupPlayers(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: invalid location: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}

		s, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, home, location).
			WillReturnError(sql.ErrNoRows)

		_, err := s.update(context.Background(), id, req)

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
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, home, location).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update player: invalid argument: the given home or location does not exist: " +
			"home '00000000-0000-0000-0000-000000000001', location '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, home, location).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := s.update(context.Background(), id, req)

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
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, home, location).
			WillReturnRows(row)

		_, err := s.update(context.Background(), id, req)

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
		req := playerRequest{Name: name, Description: description, Home: home, Location: location}
		row := sqlmock.NewRows([]string{"player_id", "name", "description", "home", "location", "created", "updated"}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupPlayers(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, home, location).
			WillReturnRows(row)

		p, err := s.update(context.Background(), id, req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if p.ID() != id ||
			p.Name() != name ||
			p.Description() != description ||
			p.Home() != home ||
			p.Location() != location {
			t.Errorf("\nExpected player: %+v", p)
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

	t.Run("invalid player id", func(t *testing.T) {
		s, _ := setupPlayers(t)

		err := s.remove(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove player: invalid argument: invalid player id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, mock := setupPlayers(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		err := s.remove(context.Background(), id)

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
		s, mock := setupPlayers(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(errors.New("unknown error"))

		err := s.remove(context.Background(), id)

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
		s, mock := setupPlayers(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := s.remove(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected err: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func setupPlayers(t *testing.T) (Players, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create sqlmock db")
	}

	return Players{DB: db}, mock
}
