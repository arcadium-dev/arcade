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

	t.Run("test player response json encoding", func(t *testing.T) {
		p := arcade.PlayerResponse{
			Data: arcade.Player{
				ID:          id,
				Name:        name,
				Description: description,
				HomeID:      homeID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.PlayerResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		player := resp.Data
		if player.ID != id ||
			player.Name != name ||
			player.Description != description ||
			player.HomeID != homeID ||
			player.LocationID != locationID {
			t.Errorf("\n%+v\n%+v", p, player)
		}
	})

	t.Run("test players response json encoding", func(t *testing.T) {
		p := arcade.PlayersResponse{
			Data: []arcade.Player{
				{
					ID:          id,
					Name:        name,
					Description: description,
					HomeID:      homeID,
					LocationID:  locationID,
					Created:     created,
					Updated:     updated,
				},
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.PlayersResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if len(resp.Data) != 1 {
			t.Fatal("sigh")
		}
		player := resp.Data[0]
		if player.ID != id ||
			player.Name != name ||
			player.Description != description ||
			player.HomeID != homeID ||
			player.LocationID != locationID {
			t.Errorf("\n%+v\n%+v", p, player)
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
		expected := "bad request: empty player name"
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
		expected := "bad request: player name exceeds maximum length"
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
		expected := "bad request: empty player description"
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
		expected := "bad request: player description exceeds maximum length"
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
		expected := "bad request: invalid homeID: '42'"
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
		expected := "bad request: invalid locationID: '42'"
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

func TestNewPlayersFilter(t *testing.T) {
	t.Run("location bad uuid", func(t *testing.T) {
		q := "locationID=42"
		_, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err == nil {
			t.Fatal("Expected an error")
		}
		expected := "bad request: invalid locationID query parameter: '42'"
		if err.Error() != expected {
			t.Errorf("\nExpected error: %s\nActual error:   %s", expected, err)
		}
	})

	t.Run("valid uuid", func(t *testing.T) {
		id := uuid.New()
		q := "locationID=" + id.String()
		filter, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.LocationID == nil {
			t.Fatal("Expected a filter locationID")
		}
		if *filter.LocationID != id {
			t.Errorf("Unexpected locationID: %s", filter.LocationID)
		}
		if filter.Limit != arcade.DefaultPlayersFilterLimit {
			t.Errorf("Unexpected limit: %d", filter.Limit)
		}
		if filter.Offset != 0 {
			t.Errorf("Unexpected offset: %d", filter.Offset)
		}
	})

	t.Run("negative limit", func(t *testing.T) {
		q := "limit=-100"
		_, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
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
		_, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
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
		_, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
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
		filter, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.LocationID != nil {
			t.Errorf("Unexpected locationID: %s", *filter.LocationID)
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
		_, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
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
		_, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
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
		filter, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: q}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.LocationID != nil {
			t.Errorf("Unexpected locationID: %s", *filter.LocationID)
		}
		if filter.Limit != arcade.DefaultPlayersFilterLimit {
			t.Errorf("Unexpected limit: %d", filter.Limit)
		}
		if filter.Offset != offset {
			t.Errorf("Unexpected offset: %d", filter.Offset)
		}
	})

	t.Run("no query parameters", func(t *testing.T) {
		filter, err := arcade.NewPlayersFilter(&http.Request{URL: &url.URL{RawQuery: ""}})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if filter.LocationID != nil {
			t.Errorf("Unexpected locationID: %s", filter.LocationID)
		}
		if filter.Limit != arcade.DefaultPlayersFilterLimit {
			t.Errorf("Unexpected limit: %d", filter.Limit)
		}
		if filter.Offset != 0 {
			t.Errorf("Unexpected offset: %d", filter.Offset)
		}
	})
}
