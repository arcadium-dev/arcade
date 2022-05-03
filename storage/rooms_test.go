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

func TestRoomsList(t *testing.T) {
	const (
		listQ = "^SELECT room_id, name, description, owner_id, parent_id, created, updated FROM rooms$"
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		ownerID     = uuid.NewString()
		parentID    = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("sql query error", func(t *testing.T) {
		r, mock := setupRooms(t)
		mock.ExpectQuery(listQ).
			WillReturnError(errors.New("unknown error"))

		_, err := r.List(context.Background(), arcade.RoomsFilter{})

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
			"room_id", "name", "description", "owner_id", "parent_id", "created", "updated",
		}).
			AddRow(id, name, description, ownerID, parentID, created, updated).
			RowError(0, errors.New("scan error"))

		r, mock := setupRooms(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		_, err := r.List(context.Background(), arcade.RoomsFilter{})

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
		rows := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated)

		r, mock := setupRooms(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		rooms, err := r.List(context.Background(), arcade.RoomsFilter{})

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(rooms) != 1 {
			t.Fatalf("Unexpected length of room list")
		}
		if rooms[0].ID != id ||
			rooms[0].Name != name ||
			rooms[0].Description != description ||
			rooms[0].OwnerID != ownerID ||
			rooms[0].ParentID != parentID {
			t.Errorf("\nExpected room: %+v", rooms[0])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestRoomsGet(t *testing.T) {
	const (
		getQ = "^SELECT room_id, name, description, owner_id, parent_id, created, updated FROM rooms WHERE room_id = (.+)$"
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		ownerID     = uuid.NewString()
		parentID    = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("invalid roomID", func(t *testing.T) {
		r, _ := setupRooms(t)

		_, err := r.Get(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get room: invalid argument: invalid room id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r, mock := setupRooms(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(sql.ErrNoRows)

		_, err := r.Get(context.Background(), id)

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
		r, mock := setupRooms(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(errors.New("unknown error"))

		_, err := r.Get(context.Background(), id)

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
		rows := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated)

		r, mock := setupRooms(t)
		mock.ExpectQuery(getQ).WillReturnRows(rows)

		room, err := r.Get(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if room.ID != id ||
			room.Name != name ||
			room.Description != description ||
			room.OwnerID != ownerID ||
			room.ParentID != parentID {
			t.Errorf("\nExpected room: %+v", room)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestRoomsCreate(t *testing.T) {
	const (
		createQ = `^INSERT INTO rooms \(name, description, owner_id, parent_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING room_id, name, description, owner_id, parent_id, created, updated$`
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		ownerID     = "00000000-0000-0000-0000-000000000001"
		parentID    = "00000000-0000-0000-0000-000000000001"
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("empty name", func(t *testing.T) {
		req := arcade.RoomRequest{Description: description, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Create(context.Background(), req)

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
		for i := 0; i <= arcade.MaxRoomNameLen; i++ {
			n += "a"
		}
		req := arcade.RoomRequest{Name: n, Description: description, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: room name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Create(context.Background(), req)

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
		for i := 0; i <= arcade.MaxRoomDescriptionLen; i++ {
			d += "a"
		}
		req := arcade.RoomRequest{Name: name, Description: d, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: room description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid ownerID", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: "42", ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: invalid ownerID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid parent", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: "42"}

		r, _ := setupRooms(t)

		_, err := r.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: invalid parentID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated)

		r, mock := setupRooms(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, parentID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := r.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create room: invalid argument: the given ownerID or parentID does not exist: " +
			"ownerID '00000000-0000-0000-0000-000000000001', parentID '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated)

		r, mock := setupRooms(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, parentID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := r.Create(context.Background(), req)

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
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated).
			RowError(0, errors.New("scan error"))

		r, mock := setupRooms(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, parentID).
			WillReturnRows(row)

		_, err := r.Create(context.Background(), req)

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
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated)

		r, mock := setupRooms(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, parentID).
			WillReturnRows(row)

		room, err := r.Create(context.Background(), req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if room.ID != id ||
			room.Name != name ||
			room.Description != description ||
			room.OwnerID != ownerID ||
			room.ParentID != parentID {
			t.Errorf("\nExpected room: %+v", room)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestRoomsUpdate(t *testing.T) {
	const (
		// updateQ = `^UPDATE rooms SET (.+) WHERE (.+) RETURNING (.+)$`
		updateQ = `^UPDATE rooms SET name = (.+), description = (.+), owner_id = (.+), parent_id = (.+) ` +
			`WHERE room_id = (.+) ` +
			`RETURNING room_id, name, description, owner_id, parent_id, created, updated$`
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		ownerID     = "00000000-0000-0000-0000-000000000001"
		parentID    = "00000000-0000-0000-0000-000000000001"
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("invalid room id", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Update(context.Background(), "42", req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: invalid room id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := arcade.RoomRequest{Description: description, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Update(context.Background(), id, req)

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
		for i := 0; i <= arcade.MaxRoomNameLen; i++ {
			n += "a"
		}
		req := arcade.RoomRequest{Name: n, Description: description, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: room name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Update(context.Background(), id, req)

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
		for i := 0; i <= arcade.MaxRoomDescriptionLen; i++ {
			d += "a"
		}
		req := arcade.RoomRequest{Name: name, Description: d, OwnerID: ownerID, ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: room description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid owner", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: "42", ParentID: parentID}

		r, _ := setupRooms(t)

		_, err := r.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: invalid ownerID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid parent", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: "42"}

		r, _ := setupRooms(t)

		_, err := r.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: invalid parentID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}

		r, mock := setupRooms(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, parentID).
			WillReturnError(sql.ErrNoRows)

		_, err := r.Update(context.Background(), id, req)

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
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated)

		r, mock := setupRooms(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, parentID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := r.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update room: invalid argument: the given ownerID or parentID does not exist: " +
			"ownerID '00000000-0000-0000-0000-000000000001', parentID '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated)

		r, mock := setupRooms(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, parentID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := r.Update(context.Background(), id, req)

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
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated).
			RowError(0, errors.New("scan error"))

		r, mock := setupRooms(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, parentID).
			WillReturnRows(row)

		_, err := r.Update(context.Background(), id, req)

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
		req := arcade.RoomRequest{Name: name, Description: description, OwnerID: ownerID, ParentID: parentID}
		row := sqlmock.NewRows([]string{"room_id", "name", "description", "owner_id", "parent_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, parentID, created, updated)

		r, mock := setupRooms(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, parentID).
			WillReturnRows(row)

		room, err := r.Update(context.Background(), id, req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if room.ID != id ||
			room.Name != name ||
			room.Description != description ||
			room.OwnerID != ownerID ||
			room.ParentID != parentID {
			t.Errorf("\nExpected room: %+v", room)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestRoomsRemove(t *testing.T) {
	const (
		removeQ = `^DELETE FROM rooms WHERE room_id = (.+)$`
	)

	var (
		id = uuid.NewString()
	)

	t.Run("invalid room id", func(t *testing.T) {
		r, _ := setupRooms(t)

		err := r.Remove(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove room: invalid argument: invalid room id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r, mock := setupRooms(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		err := r.Remove(context.Background(), id)

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
		r, mock := setupRooms(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(errors.New("unknown error"))

		err := r.Remove(context.Background(), id)

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
		r, mock := setupRooms(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := r.Remove(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected err: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func setupRooms(t *testing.T) (storage.Rooms, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create sqlmock db")
	}

	return storage.Rooms{DB: db, Driver: cockroach.Driver{}}, mock
}
