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

func TestItemsList(t *testing.T) {
	const (
		listQ = "^SELECT item_id, name, description, owner_id, location_id, inventory_id, created, updated FROM items$"
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		ownerID     = uuid.NewString()
		locationID  = uuid.NewString()
		inventoryID = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("sql query error", func(t *testing.T) {
		l, mock := setupItems(t)
		mock.ExpectQuery(listQ).
			WillReturnError(errors.New("unknown error"))

		_, err := l.List(context.Background(), arcade.ItemsFilter{})

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to list items: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("sql scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated",
		}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated).
			RowError(0, errors.New("scan error"))

		l, mock := setupItems(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		_, err := l.List(context.Background(), arcade.ItemsFilter{})

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to list items: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated)

		l, mock := setupItems(t)
		mock.ExpectQuery(listQ).
			WillReturnRows(rows).
			RowsWillBeClosed()

		items, err := l.List(context.Background(), arcade.ItemsFilter{})

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Unexpected length of item list")
		}
		if items[0].ID != id ||
			items[0].Name != name ||
			items[0].Description != description ||
			items[0].OwnerID != ownerID ||
			items[0].LocationID != locationID ||
			items[0].InventoryID != inventoryID {
			t.Errorf("\nExpected item: %+v", items[0])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestItemsGet(t *testing.T) {
	const (
		getQ = "^SELECT item_id, name, description, owner_id, location_id, inventory_id, created, updated FROM items WHERE item_id = (.+)$"
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		ownerID     = uuid.NewString()
		locationID  = uuid.NewString()
		inventoryID = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("invalid itemID", func(t *testing.T) {
		l, _ := setupItems(t)

		_, err := l.Get(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get item: invalid argument: invalid item id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		l, mock := setupItems(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(sql.ErrNoRows)

		_, err := l.Get(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get item: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		l, mock := setupItems(t)
		mock.ExpectQuery(getQ).WithArgs(id).WillReturnError(errors.New("unknown error"))

		_, err := l.Get(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to get item: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated)

		l, mock := setupItems(t)
		mock.ExpectQuery(getQ).WillReturnRows(rows)

		item, err := l.Get(context.Background(), id)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if item.ID != id ||
			item.Name != name ||
			item.Description != description ||
			item.OwnerID != ownerID ||
			item.LocationID != locationID ||
			item.InventoryID != inventoryID {
			t.Errorf("\nExpected item: %+v", item)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestItemsCreate(t *testing.T) {
	const (
		createQ = `^INSERT INTO items \(name, description, owner_id, location_id, inventory_id\) ` +
			`VALUES \((.+), (.+), (.+), (.+)\) ` +
			`RETURNING item_id, name, description, owner_id, location_id, inventory_id, created, updated$`
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		ownerID     = "00000000-0000-0000-0000-000000000001"
		locationID  = "00000000-0000-0000-0000-000000000001"
		inventoryID = "00000000-0000-0000-0000-000000000001"
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("empty name", func(t *testing.T) {
		req := arcade.ItemRequest{Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: invalid argument: empty item name"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long name", func(t *testing.T) {
		n := ""
		for i := 0; i <= arcade.MaxItemNameLen; i++ {
			n += "a"
		}
		req := arcade.ItemRequest{Name: n, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: invalid argument: item name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: invalid argument: empty item description"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long description", func(t *testing.T) {
		d := ""
		for i := 0; i <= arcade.MaxItemDescriptionLen; i++ {
			d += "a"
		}
		req := arcade.ItemRequest{Name: name, Description: d, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: invalid argument: item description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid ownerID", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: "42", LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: invalid argument: invalid ownerID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid locationID", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: "42", InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: invalid argument: invalid locationID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid inventoryID", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: "42"}

		l, _ := setupItems(t)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: invalid argument: invalid inventoryID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}
		row := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated)

		l, mock := setupItems(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, locationID, inventoryID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: invalid argument: the given ownerID, locationID, or inventoryID does not exist: " +
			"ownerID '00000000-0000-0000-0000-000000000001', locationID '00000000-0000-0000-0000-000000000001', inventoryID '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}
		row := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated)

		l, mock := setupItems(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, locationID, inventoryID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: already exists: item already exists"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}
		row := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated).
			RowError(0, errors.New("scan error"))

		l, mock := setupItems(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, locationID, inventoryID).
			WillReturnRows(row)

		_, err := l.Create(context.Background(), req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to create item: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}
		row := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated)

		l, mock := setupItems(t)
		mock.ExpectQuery(createQ).
			WithArgs(name, description, ownerID, locationID, inventoryID).
			WillReturnRows(row)

		item, err := l.Create(context.Background(), req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if item.ID != id ||
			item.Name != name ||
			item.Description != description ||
			item.OwnerID != ownerID ||
			item.LocationID != locationID ||
			item.InventoryID != inventoryID {
			t.Errorf("\nExpected item: %+v", item)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestItemsUpdate(t *testing.T) {
	const (
		// updateQ = `^UPDATE items SET (.+) WHERE (.+) RETURNING (.+)$`
		updateQ = `^UPDATE items SET name = (.+), description = (.+), owner_id = (.+), location_id = (.+), inventory_id = (.+) ` +
			`WHERE item_id = (.+) ` +
			`RETURNING item_id, name, description, owner_id, location_id, inventory_id, created, updated$`
	)

	var (
		id          = uuid.NewString()
		name        = "Nobody"
		description = "No one of importance."
		ownerID     = "00000000-0000-0000-0000-000000000001"
		locationID  = "00000000-0000-0000-0000-000000000001"
		inventoryID = "00000000-0000-0000-0000-000000000001"
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("invalid item id", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Update(context.Background(), "42", req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: invalid item id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := arcade.ItemRequest{Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: empty item name"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long name", func(t *testing.T) {
		n := ""
		for i := 0; i <= arcade.MaxItemNameLen; i++ {
			n += "a"
		}
		req := arcade.ItemRequest{Name: n, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: item name exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: empty item description"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("long description", func(t *testing.T) {
		d := ""
		for i := 0; i <= arcade.MaxItemDescriptionLen; i++ {
			d += "a"
		}
		req := arcade.ItemRequest{Name: name, Description: d, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: item description exceeds maximum length"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid ownerID", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: "42", LocationID: locationID, InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: invalid ownerID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid locationID", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: "42", InventoryID: inventoryID}

		l, _ := setupItems(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: invalid locationID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("invalid inventoryID", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: "42"}

		l, _ := setupItems(t)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: invalid inventoryID: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}

		l, mock := setupItems(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, inventoryID).
			WillReturnError(sql.ErrNoRows)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("foreign key voilation", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}
		row := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated)

		l, mock := setupItems(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, inventoryID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: invalid argument: the given ownerID, locationID, or inventoryID does not exist: " +
			"ownerID '00000000-0000-0000-0000-000000000001', locationID '00000000-0000-0000-0000-000000000001', inventoryID '00000000-0000-0000-0000-000000000001'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}
		row := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated)

		l, mock := setupItems(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, inventoryID).
			WillReturnRows(row).
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: already exists: item name is not unique"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}
		row := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated).
			RowError(0, errors.New("scan error"))

		l, mock := setupItems(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, inventoryID).
			WillReturnRows(row)

		_, err := l.Update(context.Background(), id, req)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to update item: internal error: scan error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := arcade.ItemRequest{Name: name, Description: description, OwnerID: ownerID, LocationID: locationID, InventoryID: inventoryID}
		row := sqlmock.NewRows([]string{"item_id", "name", "description", "owner_id", "location_id", "inventory_id", "created", "updated"}).
			AddRow(id, name, description, ownerID, locationID, inventoryID, created, updated)

		l, mock := setupItems(t)
		mock.ExpectQuery(updateQ).
			WithArgs(id, name, description, ownerID, locationID, inventoryID).
			WillReturnRows(row)

		item, err := l.Update(context.Background(), id, req)

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if item.ID != id ||
			item.Name != name ||
			item.Description != description ||
			item.OwnerID != ownerID ||
			item.LocationID != locationID ||
			item.InventoryID != inventoryID {
			t.Errorf("\nExpected item: %+v", item)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})
}

func TestItemsRemove(t *testing.T) {
	const (
		removeQ = `^DELETE FROM items WHERE item_id = (.+)$`
	)

	var (
		id = uuid.NewString()
	)

	t.Run("invalid item id", func(t *testing.T) {
		l, _ := setupItems(t)

		err := l.Remove(context.Background(), "42")

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove item: invalid argument: invalid item id: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		l, mock := setupItems(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		err := l.Remove(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove item: not found"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		l, mock := setupItems(t)
		mock.ExpectExec(removeQ).
			WithArgs(id).
			WillReturnError(errors.New("unknown error"))

		err := l.Remove(context.Background(), id)

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "failed to remove item: internal error: unknown error"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unexpected err: %s", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		l, mock := setupItems(t)
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

func setupItems(t *testing.T) (storage.Items, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create sqlmock db")
	}

	return storage.Items{DB: db, Driver: cockroach.Driver{}}, mock
}
