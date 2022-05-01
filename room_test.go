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
	"testing"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/arcade"
)

func TestRoomRequestValidate(t *testing.T) {
	t.Run("test empty name", func(t *testing.T) {
		r := arcade.RoomRequest{}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: empty room name"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test name length", func(t *testing.T) {
		r := arcade.RoomRequest{Name: randString(arcade.MaxRoomNameLen + 1)}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: room name exceeds maximum length"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test empty description", func(t *testing.T) {
		r := arcade.RoomRequest{
			Name: randString(42),
		}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: empty room description"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test description length", func(t *testing.T) {
		r := arcade.RoomRequest{
			Name:        randString(42),
			Description: randString(arcade.MaxRoomDescriptionLen + 1),
		}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: room description exceeds maximum length"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid ownerID", func(t *testing.T) {
		r := arcade.RoomRequest{
			Name:        randString(42),
			Description: randString(128),
			OwnerID:     "42",
		}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: invalid ownerID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid parentID", func(t *testing.T) {
		r := arcade.RoomRequest{
			Name:        randString(42),
			Description: randString(128),
			OwnerID:     uuid.NewString(),
			ParentID:    "42",
		}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: invalid parentID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		r := arcade.RoomRequest{
			Name:        randString(73),
			Description: randString(256),
			OwnerID:     uuid.NewString(),
			ParentID:    uuid.NewString(),
		}

		_, _, err := r.Validate()

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
	})
}

func TestNewRoomsReponse(t *testing.T) {
	var (
		id          = uuid.NewString()
		name        = randString(21)
		description = randString(49)
		ownerID     = uuid.NewString()
		parentID    = uuid.NewString()

		created = time.Now()
		updated = time.Now()

		p = arcade.Room{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			ParentID:    parentID,
			Created:     created,
			Updated:     updated,
		}
	)
	r := arcade.NewRoomsResponse([]arcade.Room{p})

	if r.Data[0].ID != id ||
		r.Data[0].Name != name ||
		r.Data[0].Description != description ||
		r.Data[0].OwnerID != ownerID ||
		r.Data[0].ParentID != parentID ||
		!created.Equal(r.Data[0].Created) ||
		!updated.Equal(r.Data[0].Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}
