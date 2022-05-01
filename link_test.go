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

func TestLinkRequestValidate(t *testing.T) {
	t.Run("test empty name", func(t *testing.T) {
		r := arcade.LinkRequest{}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: empty link name"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test name length", func(t *testing.T) {
		r := arcade.LinkRequest{Name: randString(arcade.MaxLinkNameLen + 1)}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: link name exceeds maximum length"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test empty description", func(t *testing.T) {
		r := arcade.LinkRequest{
			Name: randString(42),
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: empty link description"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test description length", func(t *testing.T) {
		r := arcade.LinkRequest{
			Name:        randString(42),
			Description: randString(arcade.MaxLinkDescriptionLen + 1),
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: link description exceeds maximum length"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid ownerID", func(t *testing.T) {
		r := arcade.LinkRequest{
			Name:        randString(42),
			Description: randString(128),
			OwnerID:     "42",
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: invalid ownerID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid locationID", func(t *testing.T) {
		r := arcade.LinkRequest{
			Name:        randString(42),
			Description: randString(128),
			OwnerID:     uuid.NewString(),
			LocationID:  "42",
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: invalid locationID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("test invalid destinationID", func(t *testing.T) {
		r := arcade.LinkRequest{
			Name:          randString(42),
			Description:   randString(128),
			OwnerID:       uuid.NewString(),
			LocationID:    uuid.NewString(),
			DestinationID: "42",
		}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "invalid argument: invalid destinationID: '42'"
		if expected != err.Error() {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		r := arcade.LinkRequest{
			Name:          randString(73),
			Description:   randString(256),
			OwnerID:       uuid.NewString(),
			LocationID:    uuid.NewString(),
			DestinationID: uuid.NewString(),
		}

		_, _, _, err := r.Validate()

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
	})
}

func TestNewLinksReponse(t *testing.T) {
	var (
		id          = uuid.NewString()
		name        = randString(21)
		description = randString(49)
		ownerID     = uuid.NewString()
		locationID  = uuid.NewString()

		created = time.Now()
		updated = time.Now()

		p = arcade.Link{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			LocationID:  locationID,
			Created:     created,
			Updated:     updated,
		}
	)
	r := arcade.NewLinksResponse([]arcade.Link{p})

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
