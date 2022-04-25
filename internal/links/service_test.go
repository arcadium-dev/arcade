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

package links

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
	if s.Name() != "links" {
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
		listQ = "^SELECT link_id, name, description, owner, location, destination, created, updated FROM links$"
	)

	t.Run("sql query error", func(t *testing.T) {
		s, mock := setupService(t)
		mock.ExpectQuery(listQ).
			WillReturnError(errors.New("unknown error"))

		_, err := s.list(context.Background())

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to list links: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("sql scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"link_id", "name", "description", "owner", "location", "destination", "created", "updated",
		}).
			AddRow(id, name, description, owner, location, destination, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupService(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		_, err := s.list(context.Background())

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to list links: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		links, err := s.list(context.Background())

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(links) != 1 {
			t.Fatalf("Unexpected length of link list")
		}
		if links[0].ID() != id ||
			links[0].Name() != name ||
			links[0].Description() != description ||
			links[0].Owner() != owner ||
			links[0].Location() != location ||
			links[0].Destination() != destination {
			t.Errorf("\nExpected link: %+v", links[0])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestServiceGet(t *testing.T) {
	const (
		getQ = "^SELECT link_id, name, description, owner, location, destination, created, updated FROM links WHERE link_id = (.+)$"
	)

	t.Run("invalid linkID", func(t *testing.T) {
		s, _ := setupService(t)

		_, err := s.get(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get link: invalid argument: invalid link id: '42'"
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
		expected := "failed to get link: not found"
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
		expected := "failed to get link: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated)

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
			p.Location() != location ||
			p.Destination() != destination {
			t.Errorf("\nExpected link: %+v", p)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestServiceCreate(t *testing.T) {
	const (
		createQ = `^INSERT INTO links \(name, description, owner, location, destination\) ` +
			`VALUES \((.+), (.+), (.+), (.+) (.+)\) ` +
			`RETURNING link_id, name, description, owner, location, destination, created, updated$`
	)

	t.Run("empty name", func(t *testing.T) {
		req := linkRequest{Description: description, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: empty link name"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long name", func(t *testing.T) {
		n := ""
		for i := 0; i <= maxNameLen; i++ {
			n += "a"
		}
		req := linkRequest{Name: n, Description: description, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: link name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := linkRequest{Name: name, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: empty link description"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long description", func(t *testing.T) {
		d := ""
		for i := 0; i <= maxDescriptionLen; i++ {
			d += "a"
		}
		req := linkRequest{Name: name, Description: d, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: link description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid owner", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: "42", Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: invalid owner: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid location", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: "42", Destination: destination}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: invalid location: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid destination", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: "42"}

		s, _ := setupService(t)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: invalid destination: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, owner, location, destination).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: the given owner, location, or destination does not exist: " +
			"owner '00000000-0000-0000-0000-000000000001', " +
			"location '00000000-0000-0000-0000-000000000001', " +
			"destination '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, owner, location, destination).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: already exists: link already exists"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupService(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, owner, location, destination).
			WillReturnRows(row)

		_, err := s.create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, owner, location, destination).
			WillReturnRows(row)

		_, err := s.create(context.Background(), req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if p.ID() != id ||
			p.Name() != name ||
			p.Description() != description ||
			p.Owner() != owner ||
			p.Location() != location ||
			p.Destination() != destination {
			t.Errorf("\nExpected link: %+v", p)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestServiceUpdate(t *testing.T) {
	const (
		// updateQ = `^UPDATE links SET (.+) WHERE (.+) RETURNING (.+)$`
		updateQ = `^UPDATE links SET name = (.+), description = (.+), owner = (.+), location = (.+), destination = (.+) ` +
			`WHERE link_id = (.+) ` +
			`RETURNING link_id, name, description, owner, location, destination, created, updated$`
	)

	t.Run("invalid link id", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), "42", req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: invalid link id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := linkRequest{Description: description, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: empty link name"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long name", func(t *testing.T) {
		n := ""
		for i := 0; i <= maxNameLen; i++ {
			n += "a"
		}
		req := linkRequest{Name: n, Description: description, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: link name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := linkRequest{Name: name, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: empty link description"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long description", func(t *testing.T) {
		d := ""
		for i := 0; i <= maxDescriptionLen; i++ {
			d += "a"
		}
		req := linkRequest{Name: name, Description: d, Owner: owner, Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: link description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid owner", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: "42", Location: location, Destination: destination}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: invalid owner: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid location", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: "42", Destination: destination}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: invalid location: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid destination", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: "42"}

		s, _ := setupService(t)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: invalid destination: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, location, destination).
			WillReturnError(sql.ErrNoRows)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, location, destination).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: the given owner, location, or destination does not exist: " +
			"owner '00000000-0000-0000-0000-000000000001', " +
			"location '00000000-0000-0000-0000-000000000001', " +
			"destination '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, location, destination).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: already exists: link name is not unique"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated).
			RowError(0, errors.New("scan error"))

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, location, destination).
			WillReturnRows(row)

		_, err := s.update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := linkRequest{Name: name, Description: description, Owner: owner, Location: location, Destination: destination}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner", "location", "destination", "created", "updated"}).
			AddRow(id, name, description, owner, location, destination, created, updated)

		s, mock := setupService(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, owner, location, destination).
			WillReturnRows(row)

		p, err := s.update(context.Background(), id, req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if p.ID() != id ||
			p.Name() != name ||
			p.Description() != description ||
			p.Owner() != owner ||
			p.Location() != location ||
			p.Destination() != destination {
			t.Errorf("\nExpected link: %+v", p)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestServiceRemove(t *testing.T) {
	const (
		removeQ = `^DELETE FROM links WHERE link_id = (.+)$`
	)

	t.Run("invalid link id", func(t *testing.T) {
		s, _ := setupService(t)

		err := s.remove(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove link: invalid argument: invalid link id: '42'"
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
		expected := "failed to remove link: not found"
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
		expected := "failed to remove link: internal error: unknown error"
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
