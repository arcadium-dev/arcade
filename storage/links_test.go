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

package storage_test

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

func TestLinksList(t *testing.T) {
	const (
		listQ = "^SELECT link_id, name, description, owner_id, location_id, destination_id, created, updated FROM links$"
	)

	var (
		id            = uuid.NewString()
		name          = "Nobody"
		description   = "No one of importance."
		ownerID       = uuid.NewString()
		locationID    = uuid.NewString()
		destinationID = uuid.NewString()
		created       = time.Now()
		updated       = time.Now()
	)

	t.Run("sql query error", func(t *testing.T) {
		l, mock := setupLinks(t)
		mock.ExpectQuery(listQ).
			WillReturnError(errors.New("unknown error"))

		_, err := l.List(context.Background(), arcade.LinksFilter{})

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
			"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated",
		}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated).
			RowError(0, errors.New("scan error"))

		l, mock := setupLinks(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		_, err := l.List(context.Background(), arcade.LinksFilter{})

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
		rows := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated)

		l, mock := setupLinks(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		links, err := l.List(context.Background(), arcade.LinksFilter{})

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(links) != 1 {
			t.Fatalf("Unexpected length of link list")
		}
		if links[0].ID != id ||
			links[0].Name != name ||
			links[0].Description != description ||
			links[0].OwnerID != ownerID ||
			links[0].LocationID != locationID ||
			links[0].DestinationID != destinationID {
			t.Errorf("\nExpected link: %+v", links[0])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestLinksGet(t *testing.T) {
	const (
		getQ = "^SELECT link_id, name, description, owner_id, location_id, destination_id, created, updated FROM links WHERE link_id = (.+)$"
	)

	var (
		id            = uuid.NewString()
		name          = "Nobody"
		description   = "No one of importance."
		ownerID       = uuid.NewString()
		locationID    = uuid.NewString()
		destinationID = uuid.NewString()
		created       = time.Now()
		updated       = time.Now()
	)

	t.Run("invalid linkID", func(t *testing.T) {
		l, _ := setupLinks(t)

		_, err := l.Get(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get link: invalid argument: invalid link id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		l, mock := setupLinks(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(sql.ErrNoRows)

		_, err := l.Get(context.Background(), id)

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
		l, mock := setupLinks(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(errors.New("unknown error"))

		_, err := l.Get(context.Background(), id)

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
		rows := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated)

		l, mock := setupLinks(t)
		mock.ExpectQuery(getQ).WillReturnRows(rows)

		link, err := l.Get(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if link.ID != id ||
			link.Name != name ||
			link.Description != description ||
			link.OwnerID != ownerID ||
			link.LocationID != locationID ||
			link.DestinationID != destinationID {
			t.Errorf("\nExpected link: %+v", link)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestLinksCreate(t *testing.T) {
	const (
		createQ = `^INSERT INTO links \(name, description, owner_id, location_id, destination_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING link_id, name, description, owner_id, location_id, destination_id, created, updated$`
	)

	var (
		id            = uuid.NewString()
		name          = "Nobody"
		description   = "No one of importance."
		ownerID       = "00000000-0000-0000-0000-000000000001"
		locationID    = "00000000-0000-0000-0000-000000000001"
		destinationID = "00000000-0000-0000-0000-000000000001"
		created       = time.Now()
		updated       = time.Now()
	)

	t.Run("empty name", func(t *testing.T) {
		req := arcade.LinkRequest{Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Create(context.Background(), req)

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
		for i := 0; i <= arcade.MaxLinkNameLen; i++ {
			n += "a"
		}
		req := arcade.LinkRequest{Name: n, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: link name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Create(context.Background(), req)

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
		for i := 0; i <= arcade.MaxLinkDescriptionLen; i++ {
			d += "a"
		}
		req := arcade.LinkRequest{Name: name, Description: d, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: link description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid ownerID", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: "42", LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: invalid ownerID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid locationID", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: "42", DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: invalid locationID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid destinationID", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: "42"}

		l, _ := setupLinks(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: invalid destinationID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated)

		l, mock := setupLinks(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, locationID, destinationID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create link: invalid argument: the given ownerID, locationID, or destinationID does not exist: " +
			"ownerID '00000000-0000-0000-0000-000000000001', locationID '00000000-0000-0000-0000-000000000001', destinationID '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated)

		l, mock := setupLinks(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, locationID, destinationID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := l.Create(context.Background(), req)

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
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated).
			RowError(0, errors.New("scan error"))

		l, mock := setupLinks(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, locationID, destinationID).
			WillReturnRows(row)

		_, err := l.Create(context.Background(), req)

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
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated)

		l, mock := setupLinks(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, locationID, destinationID).
			WillReturnRows(row)

		link, err := l.Create(context.Background(), req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if link.ID != id ||
			link.Name != name ||
			link.Description != description ||
			link.OwnerID != ownerID ||
			link.LocationID != locationID ||
			link.DestinationID != destinationID {
			t.Errorf("\nExpected link: %+v", link)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestLinksUpdate(t *testing.T) {
	const (
		// updateQ = `^UPDATE links SET (.+) WHERE (.+) RETURNING (.+)$`
		updateQ = `^UPDATE links SET name = (.+), description = (.+), owner_id = (.+), location_id = (.+), destination_id = (.+) ` +
			`WHERE link_id = (.+) ` +
			`RETURNING link_id, name, description, owner_id, location_id, destination_id, created, updated$`
	)

	var (
		id            = uuid.NewString()
		name          = "Nobody"
		description   = "No one of importance."
		ownerID       = "00000000-0000-0000-0000-000000000001"
		locationID    = "00000000-0000-0000-0000-000000000001"
		destinationID = "00000000-0000-0000-0000-000000000001"
		created       = time.Now()
		updated       = time.Now()
	)

	t.Run("invalid link id", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Update(context.Background(), "42", req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: invalid link id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := arcade.LinkRequest{Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Update(context.Background(), id, req)

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
		for i := 0; i <= arcade.MaxLinkNameLen; i++ {
			n += "a"
		}
		req := arcade.LinkRequest{Name: n, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: link name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Update(context.Background(), id, req)

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
		for i := 0; i <= arcade.MaxLinkDescriptionLen; i++ {
			d += "a"
		}
		req := arcade.LinkRequest{Name: name, Description: d, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: link description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid ownerID", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: "42", LocationID: locationID, DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: invalid ownerID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid locationID", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: "42", DestinationID: destinationID}

		l, _ := setupLinks(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: invalid locationID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid destinationID", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: "42"}

		l, _ := setupLinks(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: invalid destinationID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}

		l, mock := setupLinks(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, destinationID).
			WillReturnError(sql.ErrNoRows)

		_, err := l.Update(context.Background(), id, req)

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
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated)

		l, mock := setupLinks(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, destinationID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update link: invalid argument: the given ownerID, locationID, or destinationID does not exist: " +
			"ownerID '00000000-0000-0000-0000-000000000001', locationID '00000000-0000-0000-0000-000000000001', destinationID '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated)

		l, mock := setupLinks(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, destinationID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := l.Update(context.Background(), id, req)

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
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated).
			RowError(0, errors.New("scan error"))

		l, mock := setupLinks(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, destinationID).
			WillReturnRows(row)

		_, err := l.Update(context.Background(), id, req)

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
		req := arcade.LinkRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, DestinationID: destinationID}
		row := sqlmock.NewRows([]string{"link_id", "name", "description", "owner_id", "location_id", "destination_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, destinationID, created, updated)

		l, mock := setupLinks(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, destinationID).
			WillReturnRows(row)

		link, err := l.Update(context.Background(), id, req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if link.ID != id ||
			link.Name != name ||
			link.Description != description ||
			link.OwnerID != ownerID ||
			link.LocationID != locationID ||
			link.DestinationID != destinationID {
			t.Errorf("\nExpected link: %+v", link)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestLinksRemove(t *testing.T) {
	const (
		removeQ = `^DELETE FROM links WHERE link_id = (.+)$`
	)

	var (
		id = uuid.NewString()
	)

	t.Run("invalid link id", func(t *testing.T) {
		l, _ := setupLinks(t)

		err := l.Remove(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove link: invalid argument: invalid link id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		l, mock := setupLinks(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		err := l.Remove(context.Background(), id)

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
		l, mock := setupLinks(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(errors.New("unknown error"))

		err := l.Remove(context.Background(), id)

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
		l, mock := setupLinks(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := l.Remove(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected err: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func setupLinks(t *testing.T) (storage.Links, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create sqlmock db")
	}

	return storage.Links{DB: db, Driver: cockroach.Driver{}}, mock
}
