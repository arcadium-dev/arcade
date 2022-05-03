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

func TestPlayerJSONEncoding(t *testing.T) {
	var (
		id          = uuid.NewString()
		name        = randString(21)
		description = randString(49)
		homeID      = uuid.NewString()
		locationID  = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("test player json encoding", func(t *testing.T) {
		p := arcade.Player{
			ID:          id,
			Name:        name,
			Description: description,
			HomeID:      homeID,
			LocationID:  locationID,
			Created:     created,
			Updated:     updated,
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var player arcade.Player
		if err := json.Unmarshal(b, &player); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if player.ID != id ||
			player.Name != name ||
			player.Description != description ||
			player.HomeID != homeID ||
			player.LocationID != locationID {
			t.Errorf("\n%+v\n%+v", p, player)
		}
	})

	t.Run("test player request json encoding", func(t *testing.T) {
		r := arcade.PlayerRequest{
			Name:        randString(73),
			Description: randString(256),
			HomeID:      uuid.NewString(),
			LocationID:  uuid.NewString(),
		}

		b, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var req arcade.PlayerRequest
		if err := json.Unmarshal(b, &req); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if r != req {
			t.Error("bummer")
		}
	})
}

func TestPlayerRequestValidate(t *testing.T) {
	t.Run("test empty name", func(t *testing.T) {
		r := arcade.PlayerRequest{}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: empty player name"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test name length", func(t *testing.T) {
		r := arcade.PlayerRequest{Name: randString(arcade.MaxPlayerNameLen + 1)}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: player name exceeds maximum length"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test empty description", func(t *testing.T) {
		r := arcade.PlayerRequest{
			Name: randString(42),
		}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: empty player description"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test description length", func(t *testing.T) {
		r := arcade.PlayerRequest{
			Name:        randString(42),
			Description: randString(arcade.MaxPlayerDescriptionLen + 1),
		}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: player description exceeds maximum length"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid homeID", func(t *testing.T) {
		r := arcade.PlayerRequest{
			Name:        randString(42),
			Description: randString(128),
			HomeID:      "42",
		}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: invalid homeID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid locationID", func(t *testing.T) {
		r := arcade.PlayerRequest{
			Name:        randString(42),
			Description: randString(128),
			HomeID:      uuid.NewString(),
			LocationID:  "42",
		}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: invalid locationID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		r := arcade.PlayerRequest{
			Name:        randString(73),
			Description: randString(256),
			HomeID:      uuid.NewString(),
			LocationID:  uuid.NewString(),
		}

		_, _, err := r.Validate()

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
	})
}

func TestNewPlayersReponse(t *testing.T) {
	var (
		id          = uuid.NewString()
		name        = randString(21)
		description = randString(49)
		homeID      = uuid.NewString()
		locationID  = uuid.NewString()

		created = time.Now()
		updated = time.Now()

		p = arcade.Player{
			ID:          id,
			Name:        name,
			Description: description,
			HomeID:      homeID,
			LocationID:  locationID,
			Created:     created,
			Updated:     updated,
		}
	)
	r := arcade.NewPlayersResponse([]arcade.Player{p})

	if r.Data[0].ID != id ||
		r.Data[0].Name != name ||
		r.Data[0].Description != description ||
		r.Data[0].HomeID != homeID ||
		r.Data[0].LocationID != locationID ||
		!created.Equal(r.Data[0].Created) ||
		!updated.Equal(r.Data[0].Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}
