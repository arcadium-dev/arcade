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

func TestLinkJSONEncoding(t *testing.T) {
	var (
		id            = uuid.NewString()
		name          = randString(21)
		description   = randString(49)
		ownerID       = uuid.NewString()
		locationID    = uuid.NewString()
		destinationID = uuid.NewString()
		created       = time.Now()
		updated       = time.Now()
	)

	t.Run("test link json encoding", func(t *testing.T) {
		p := arcade.Link{
			ID:            id,
			Name:          name,
			Description:   description,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
			Created:       created,
			Updated:       updated,
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var link arcade.Link
		if err := json.Unmarshal(b, &link); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if link.ID != id ||
			link.Name != name ||
			link.Description != description ||
			link.OwnerID != ownerID ||
			link.LocationID != locationID ||
			link.DestinationID != destinationID {
			t.Errorf("\n%+v\n%+v", p, link)
		}
	})

	t.Run("test link request json encoding", func(t *testing.T) {
		r := arcade.LinkRequest{
			Name:          randString(73),
			Description:   randString(256),
			OwnerID:       uuid.NewString(),
			LocationID:    uuid.NewString(),
			DestinationID: uuid.NewString(),
		}

		b, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var req arcade.LinkRequest
		if err := json.Unmarshal(b, &req); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if r != req {
			t.Error("bummer")
		}
	})

	t.Run("test link response json encoding", func(t *testing.T) {
		p := arcade.LinkResponse{
			Data: arcade.Link{
				ID:            id,
				Name:          name,
				Description:   description,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
				Created:       created,
				Updated:       updated,
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.LinkResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		link := resp.Data
		if link.ID != id ||
			link.Name != name ||
			link.Description != description ||
			link.OwnerID != ownerID ||
			link.LocationID != locationID ||
			link.DestinationID != destinationID {
			t.Errorf("\n%+v\n%+v", p, link)
		}
	})

	t.Run("test links response json encoding", func(t *testing.T) {
		p := arcade.LinksResponse{
			Data: []arcade.Link{
				{
					ID:            id,
					Name:          name,
					Description:   description,
					OwnerID:       ownerID,
					LocationID:    locationID,
					DestinationID: destinationID,
					Created:       created,
					Updated:       updated,
				},
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.LinksResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(resp.Data) != 1 {
			t.Fatal("sigh")
		}
		link := resp.Data[0]
		if link.ID != id ||
			link.Name != name ||
			link.Description != description ||
			link.OwnerID != ownerID ||
			link.LocationID != locationID ||
			link.DestinationID != destinationID {
			t.Errorf("\n%+v\n%+v", p, link)
		}
	})
}

func TestLinkRequestValidate(t *testing.T) {
	t.Run("test empty name", func(t *testing.T) {
		r := arcade.LinkRequest{}

		_, _, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: empty link name"
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
		expected := "bad request: link name exceeds maximum length"
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
		expected := "bad request: empty link description"
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
		expected := "bad request: link description exceeds maximum length"
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
		expected := "bad request: invalid ownerID: '42'"
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
		expected := "bad request: invalid locationID: '42'"
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
		expected := "bad request: invalid destinationID: '42'"
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
