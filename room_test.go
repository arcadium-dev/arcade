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
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/arcade"
)

func TestRoomJSONEncoding(t *testing.T) {
	var (
		id          = uuid.NewString()
		name        = randString(21)
		description = randString(49)
		ownerID     = uuid.NewString()
		parentID    = uuid.NewString()
		created     = time.Now()
		updated     = time.Now()
	)

	t.Run("test room json encoding", func(t *testing.T) {
		p := arcade.Room{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			ParentID:    parentID,
			Created:     created,
			Updated:     updated,
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var room arcade.Room
		if err := json.Unmarshal(b, &room); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if room.ID != id ||
			room.Name != name ||
			room.Description != description ||
			room.OwnerID != ownerID ||
			room.ParentID != parentID {
			t.Errorf("\n%+v\n%+v", p, room)
		}
	})

	t.Run("test room request json encoding", func(t *testing.T) {
		r := arcade.RoomRequest{
			Name:        randString(73),
			Description: randString(256),
			OwnerID:     uuid.NewString(),
			ParentID:    uuid.NewString(),
		}

		b, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var req arcade.RoomRequest
		if err := json.Unmarshal(b, &req); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if r != req {
			t.Error("bummer")
		}
	})

	t.Run("test room response json encoding", func(t *testing.T) {
		p := arcade.RoomResponse{
			Data: arcade.Room{
				ID:          id,
				Name:        name,
				Description: description,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.RoomResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		room := resp.Data
		if room.ID != id ||
			room.Name != name ||
			room.Description != description ||
			room.OwnerID != ownerID ||
			room.ParentID != parentID {
			t.Errorf("\n%+v\n%+v", p, room)
		}
	})

	t.Run("test rooms response json encoding", func(t *testing.T) {
		p := arcade.RoomsResponse{
			Data: []arcade.Room{
				{
					ID:          id,
					Name:        name,
					Description: description,
					OwnerID:     ownerID,
					ParentID:    parentID,
					Created:     created,
					Updated:     updated,
				},
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.RoomsResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(resp.Data) != 1 {
			t.Fatal("sigh")
		}
		room := resp.Data[0]
		if room.ID != id ||
			room.Name != name ||
			room.Description != description ||
			room.OwnerID != ownerID ||
			room.ParentID != parentID {
			t.Errorf("\n%+v\n%+v", p, room)
		}
	})
}

func TestRoomRequestValidate(t *testing.T) {
	t.Run("test empty name", func(t *testing.T) {
		r := arcade.RoomRequest{}

		_, _, err := r.Validate()

		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: empty room name"
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
		expected := "bad request: room name exceeds maximum length"
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
		expected := "bad request: empty room description"
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
		expected := "bad request: room description exceeds maximum length"
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
		expected := "bad request: invalid ownerID: '42'"
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
		expected := "bad request: invalid parentID: '42'"
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

func TestNewRoomsFilter(t *testing.T) {
	t.Run("owner bad uuid", func(t *testing.T) {
		q := "ownerID=42"
		_, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid ownerID query parameter: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("valid owner uuid", func(t *testing.T) {
		id := uuid.New()
		q := "ownerID=" + id.String()
		filter, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.OwnerID == nil {
			t.Fatal("Expected a filter ownerID")
		}
		if *filter.OwnerID != id {
			t.Errorf("Unexpected ownerID: %s", filter.OwnerID)
		}
		if filter.Limit != arcade.DefaultRoomsFilterLimit {
			t.Errorf("Unexpected limit: %d", filter.Limit)
		}
		if filter.Offset != 0 {
			t.Errorf("Unexpected offset: %d", filter.Offset)
		}
	})

	t.Run("parent bad uuid", func(t *testing.T) {
		q := "parentID=42"
		_, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid parentID query parameter: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("valid parent uuid", func(t *testing.T) {
		id := uuid.New()
		q := "parentID=" + id.String()
		filter, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.ParentID == nil {
			t.Fatal("Expected a filter parentID")
		}
		if *filter.ParentID != id {
			t.Errorf("Unexpected parentID: %s", filter.ParentID)
		}
		if filter.Limit != arcade.DefaultRoomsFilterLimit {
			t.Errorf("Unexpected limit: %d", filter.Limit)
		}
		if filter.Offset != 0 {
			t.Errorf("Unexpected offset: %d", filter.Offset)
		}
	})

	t.Run("negative limit", func(t *testing.T) {
		q := "limit=-100"
		_, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid limit query parameter: '-100'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("non-number limit", func(t *testing.T) {
		q := "limit=foo"
		_, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid limit query parameter: 'foo'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("limit greater than max", func(t *testing.T) {
		q := "limit=4096"
		_, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid limit query parameter: '4096'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("valid limit", func(t *testing.T) {
		limit := 42
		q := fmt.Sprintf("limit=%d", limit)
		filter, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.OwnerID != nil {
			t.Errorf("Unexpected ownerID: %s", *filter.OwnerID)
		}
		if filter.Limit != limit {
			t.Errorf("Unexpected limit: %d", filter.Limit)
		}
		if filter.Offset != 0 {
			t.Errorf("Unexpected offset: %d", filter.Offset)
		}
	})

	t.Run("negative offset", func(t *testing.T) {
		q := "offset=-100"
		_, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid offset query parameter: '-100'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("non-number offset", func(t *testing.T) {
		q := "offset=foo"
		_, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid offset query parameter: 'foo'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("valid offset", func(t *testing.T) {
		offset := 42
		q := fmt.Sprintf("offset=%d", offset)
		filter, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.OwnerID != nil {
			t.Errorf("Unexpected ownerID: %s", *filter.OwnerID)
		}
		if filter.Limit != arcade.DefaultRoomsFilterLimit {
			t.Errorf("Unexpected limit: %d", filter.Limit)
		}
		if filter.Offset != offset {
			t.Errorf("Unexpected offset: %d", filter.Offset)
		}
	})

	t.Run("no query parameters", func(t *testing.T) {
		filter, err := arcade.NewRoomsFilter(&http.Request{URL: &url.URL{RawQuery: ""}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.OwnerID != nil {
			t.Errorf("Unexpected ownerID: %s", filter.OwnerID)
		}
		if filter.ParentID != nil {
			t.Errorf("Unexpected parentID: %s", filter.OwnerID)
		}
		if filter.Limit != arcade.DefaultRoomsFilterLimit {
			t.Errorf("Unexpected limit: %d", filter.Limit)
		}
		if filter.Offset != 0 {
			t.Errorf("Unexpected offset: %d", filter.Offset)
		}
	})
}
