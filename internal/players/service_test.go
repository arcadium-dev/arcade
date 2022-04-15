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

package players

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestServiceNew(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create mock db")
	}

	s := New(db)

	if s.db != db {
		t.Error("Failed to set service db")
	}
	if s.h.s == nil {
		t.Error("Failed to set handler service")
	}
}

func TestServiceName(t *testing.T) {
	s := Service{}
	if s.Name() != "players" {
		t.Error("Unexpected service name")
	}
}

func TestServiceShutdown(t *testing.T) {
	s := Service{}
	s.Shutdown()

	// This is a placeholder for when we have a background monitor service running.
}

func TestServiceList(t *testing.T) {
	const (
		listQ = "^SELECT player_id, name, description, home, location, created, updated FROM players$"
	)

	t.Run("sql query error", func(t *testing.T) {
		s, mock := setupService(t)
		mock.ExpectQuery(listQ).WillReturnError(errors.New("unknown error"))

		_, err := s.list(context.Background())

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "internal error: failed to list players: unknown error"
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

		s, mock := setupService(t)
		mock.ExpectQuery(listQ).WillReturnRows(rows)

		_, err := s.list(context.Background())

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "internal error: failed to list players: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"player_id", "name", "description", "home", "location", "created", "updated",
		}).
			AddRow(id, name, description, home, location, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(listQ).WillReturnRows(rows)

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
	})
}

func TestServiceGet(t *testing.T) {
}

func TestServiceCreate(t *testing.T) {
}

func TestServiceUpdate(t *testing.T) {
}

func TestServiceRemove(t *testing.T) {
}

func setupService(t *testing.T) (Service, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create sqlmock db")
	}

	return New(db), mock
}
