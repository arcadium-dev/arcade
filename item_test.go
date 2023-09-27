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
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or impliep.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package arcade_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/arcade"
)

func TestItemJSONEncoding(t *testing.T) {
	var (
		id          = uuid.NewString()
		name        = randString(21)
		description = randString(49)
		ownerID     = uuid.NewString()
		locationID  = uuid.NewString()
		inventoryID = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("test item json encoding", func(t *testing.T) {
		p := arcade.Item{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			LocationID:  locationID,
			InventoryID: inventoryID,
			Created:     created,
			Updated:     updated,
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var item arcade.Item
		if err := json.Unmarshal(b, &item); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if item.ID != id ||
			item.Name != name ||
			item.Description != description ||
			item.OwnerID != ownerID ||
			item.LocationID != locationID ||
			item.InventoryID != inventoryID {
			t.Errorf("\n%+v\n%+v", p, item)
		}
	})

	t.Run("test item request json encoding", func(t *testing.T) {
		r := arcade.ItemRequest{
			Name:        randString(73),
			Description: randString(256),
			OwnerID:     uuid.NewString(),
			LocationID:  uuid.NewString(),
			InventoryID: uuid.NewString(),
		}

		b, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var req arcade.ItemRequest
		if err := json.Unmarshal(b, &req); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if r != req {
			t.Error("bummer")
		}
	})

	t.Run("test item response json encoding", func(t *testing.T) {
		p := arcade.ItemResponse{
			Data: arcade.Item{
				ID:          id,
				Name:        name,
				Description: description,
				OwnerID:     ownerID,
				LocationID:  locationID,
				InventoryID: inventoryID,
				Created:     created,
				Updated:     updated,
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.ItemResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		item := resp.Data
		if item.ID != id ||
			item.Name != name ||
			item.Description != description ||
			item.OwnerID != ownerID ||
			item.LocationID != locationID ||
			item.InventoryID != inventoryID {
			t.Errorf("\n%+v\n%+v", p, item)
		}
	})

	t.Run("test items response json encoding", func(t *testing.T) {
		p := arcade.ItemsResponse{
			Data: []arcade.Item{
				{
					ID:          id,
					Name:        name,
					Description: description,
					OwnerID:     ownerID,
					LocationID:  locationID,
					InventoryID: inventoryID,
					Created:     created,
					Updated:     updated,
				},
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.ItemsResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(resp.Data) != 1 {
			t.Fatal("sigh")
		}
		item := resp.Data[0]
		if item.ID != id ||
			item.Name != name ||
			item.Description != description ||
			item.OwnerID != ownerID ||
			item.LocationID != locationID ||
			item.InventoryID != inventoryID {
			t.Errorf("\n%+v\n%+v", p, item)
		}
	})
}

func TestItemRequestValidate(t *testing.T) {
	t.Run("test empty name", func(t *testing.T) {
		r := arcade.ItemRequest{}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: empty item name"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test name length", func(t *testing.T) {
		r := arcade.ItemRequest{Name: randString(arcade.MaxItemNameLen + 1)}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: item name exceeds maximum length"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test empty description", func(t *testing.T) {
		r := arcade.ItemRequest{
			Name: randString(42),
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: empty item description"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test description length", func(t *testing.T) {
		r := arcade.ItemRequest{
			Name:        randString(42),
			Description: randString(arcade.MaxItemDescriptionLen + 1),
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: item description exceeds maximum length"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid ownerID", func(t *testing.T) {
		r := arcade.ItemRequest{
			Name:        randString(42),
			Description: randString(128),
			OwnerID:     "42",
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid ownerID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid locationID", func(t *testing.T) {
		r := arcade.ItemRequest{
			Name:        randString(42),
			Description: randString(128),
			OwnerID:     uuid.NewString(),
			LocationID:  "42",
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid locationID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid inventoryID", func(t *testing.T) {
		r := arcade.ItemRequest{
			Name:        randString(42),
			Description: randString(128),
			OwnerID:     uuid.NewString(),
			LocationID:  uuid.NewString(),
			InventoryID: "42",
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid inventoryID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		r := arcade.ItemRequest{
			Name:        randString(73),
			Description: randString(256),
			OwnerID:     uuid.NewString(),
			LocationID:  uuid.NewString(),
			InventoryID: uuid.NewString(),
		}

		_, _, _, err := r.Validate()

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
	})
}

func TestNewItemsReponse(t *testing.T) {
	var (
		id          = uuid.NewString()
		name        = randString(21)
		description = randString(49)
		ownerID     = uuid.NewString()
		locationID  = uuid.NewString()

		created = time.Now()
		updated = time.Now()

		p = arcade.Item{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			LocationID:  locationID,
			Created:     created,
			Updated:     updated,
		}
	)
	r := arcade.NewItemsResponse([]arcade.Item{p})

	if r.Data[0].ID != id ||
		r.Data[0].Name != name ||
		r.Data[0].Description != description ||
		r.Data[0].OwnerID != ownerID ||
		r.Data[0].LocationID != locationID ||
		!created.Equal(r.Data[0].Created) ||
		!updated.Equal(r.Data[0].Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}
