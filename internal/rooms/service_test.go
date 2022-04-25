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

package rooms

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
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
	if s.Name() != "rooms" {
		t.Error("Unexpected service name")
	}
}

func TestServiceShutdown(t *testing.T) {
	// This is a placeholder for when we have a background monitor service running.
	s := Service{}
	s.Shutdown()
}

func TestServiceList(t *testing.T) {
	const (
		listQ = "^SELECT room_id, name, description, owner, parent, created, updated FROM rooms$"
	)

	t.Run("sql query error", func(t *testing.T) {
		s, mock := setupService(t)
		mock.ExpectQuery(listQ).
			WillReturnError(errors.New("unknown error"))

		_, err := s.list(context.Background())

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to list rooms: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("sql scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"room_id", "name", "description", "owner", "parent", "created", "updated",
		}).
			AddRow(id, name, description, owner, parent, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupService(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		_, err := s.list(context.Background())

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to list rooms: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		rooms, err := s.list(context.Background())

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(rooms) != 1 {
			t.Fatalf("Unexpected length of room list")
		}
		if rooms[0].ID() != id ||
			rooms[0].Name() != name ||
			rooms[0].Description() != description ||
			rooms[0].Owner() != owner ||
			rooms[0].Parent() != parent {
			t.Errorf("\nExpected room: %+v", rooms[0])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestServiceGet(t *testing.T) {
	const (
		getQ = "^SELECT room_id, name, description, owner, parent, created, updated FROM rooms WHERE room_id = (.+)$"
	)

	t.Run("invalid roomID", func(t *testing.T) {
		s, _ := setupService(t)

		_, err := s.get(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get room: invalid argument: invalid room id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, mock := setupService(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(sql.ErrNoRows)

		_, err := s.get(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get room: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		s, mock := setupService(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(errors.New("unknown error"))

		_, err := s.get(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get room: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(getQ).WillReturnRows(rows)

		p, err := s.get(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if p.ID() != id ||
			p.Name() != name ||
			p.Description() != description ||
			p.Owner() != owner ||
			p.Parent() != parent {
			t.Errorf("\nExpected room: %+v", p)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestServiceCreate(t *testing.T) {
	const (
		createQ = `^INSERT INTO rooms \(name, description, owner, parent\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING room_id, name, description, owner, parent, created, updated$`
	)

	t.Run("empty name", func(t *testing.T) {
		req := roomRequest{Description: description, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: empty room name"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long name", func(t *testing.T) {
		n := ""
		for i := 0; i <= maxNameLen; i++ {
			n += "a"
		}
		req := roomRequest{Name: n, Description: description, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: room name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := roomRequest{Name: name, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: empty room description"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long description", func(t *testing.T) {
		d := ""
		for i := 0; i <= maxDescriptionLen; i++ {
			d += "a"
		}
		req := roomRequest{Name: name, Description: d, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: room description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid owner", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: "42", Parent: parent}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: invalid owner: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid parent", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: "42"}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: invalid parent: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, owner, parent).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: the given owner or parent given does not exist: " +
			"owner '00000000-0000-0000-0000-000000000001', parent '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, owner, parent).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: already exists: room already exists"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupService(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, owner, parent).
			WillReturnRows(row)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, owner, parent).
			WillReturnRows(row)

		_, err := s.create(context.Background(), req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if p.ID() != id ||
			p.Name() != name ||
			p.Description() != description ||
			p.Owner() != owner ||
			p.Parent() != parent {
			t.Errorf("\nExpected room: %+v", p)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestServiceUpdate(t *testing.T) {
	const (
		// updateQ = `^UPDATE rooms SET (.+) WHERE (.+) RETURNING (.+)$`
		updateQ = `^UPDATE rooms SET name = (.+), description = (.+), owner = (.+), parent = (.+) ` +
			`WHERE room_id = (.+) ` +
			`RETURNING room_id, name, description, owner, parent, created, updated$`
	)

	t.Run("invalid room id", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), "42", req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: invalid room id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := roomRequest{Description: description, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: empty room name"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long name", func(t *testing.T) {
		n := ""
		for i := 0; i <= maxNameLen; i++ {
			n += "a"
		}
		req := roomRequest{Name: n, Description: description, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: room name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := roomRequest{Name: name, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: empty room description"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long description", func(t *testing.T) {
		d := ""
		for i := 0; i <= maxDescriptionLen; i++ {
			d += "a"
		}
		req := roomRequest{Name: name, Description: d, Owner: owner, Parent: parent}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: room description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid owner", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: "42", Parent: parent}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: invalid owner: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid parent", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: "42"}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: invalid parent: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, parent).
			WillReturnError(sql.ErrNoRows)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, parent).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: the given owner or parent given does not exist: " +
			"owner '00000000-0000-0000-0000-000000000001', parent '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, parent).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: already exists: room name is not unique"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, parent).
			WillReturnRows(row)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := roomRequest{Name: name, Description: description, Owner: owner, Parent: parent}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner", "parent", "created", "updated"}).
			AddRow(id, name, description, owner, parent, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, parent).
			WillReturnRows(row)

		p, err := s.update(context.Background(), id, req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if p.ID() != id ||
			p.Name() != name ||
			p.Description() != description ||
			p.Owner() != owner ||
			p.Parent() != parent {
			t.Errorf("\nExpected room: %+v", p)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestServiceRemove(t *testing.T) {
	const (
		removeQ = `^DELETE FROM rooms WHERE room_id = (.+)$`
	)

	t.Run("invalid room id", func(t *testing.T) {
		s, _ := setupService(t)

		err := s.remove(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove room: invalid argument: invalid room id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, mock := setupService(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		err := s.remove(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove room: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		s, mock := setupService(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(errors.New("unknown error"))

		err := s.remove(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove room: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		s, mock := setupService(t)
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

func setupService(t *testing.T) (Service, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create sqlmock db")
	}

	return New(db), mock
}
