package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/service"
)

func TestPlayersServiceName(t *testing.T) {
	s := service.PlayersService{}
	if s.Name() != "players" {
		t.Error("Unexpected service name")
	}
}

func TestPlayersServiceList(t *testing.T) {
	t.Run("filter error", func(t *testing.T) {
		route := fmt.Sprintf("%s?locationID=42", service.PlayersRoute)
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodGet, route, nil),
			http.StatusBadRequest,
			"invalid argument: invalid locationID query parameter: '42'",
		)
	})

	t.Run("service error", func(t *testing.T) {
		err := errors.New("unknown error")
		m := &mockPlayersStorage{t: t, err: err}

		checkRespError(
			t, invokePlayersService(t, m, http.MethodGet, service.PlayersRoute, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.listCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		players := []arcade.Player{
			{
				ID:          "c39761fc-5096-4b1c-9d02-c75730b7b8bf",
				Name:        "Drunen",
				Description: "Son of Martin",
				HomeID:      "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				LocationID:  "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
			},
		}
		m := &mockPlayersStorage{t: t, players: players}

		w := invokePlayersService(t, m, http.MethodGet, service.PlayersRoute, nil)

		if !m.listCalled {
			t.Error("expected list to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body")
		}
		defer resp.Body.Close()

		var playersResp arcade.PlayersResponse
		err = json.Unmarshal(body, &playersResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		if len(playersResp.Data) != len(players) {
			t.Fatalf("Unexpected players response data length: %d", len(playersResp.Data))
		}

		player := playersResp.Data[0]
		if player.ID != players[0].ID ||
			player.Name != players[0].Name ||
			player.Description != players[0].Description ||
			player.HomeID != players[0].HomeID ||
			player.LocationID != players[0].LocationID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestPlayersServiceGet(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		homeID      = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID  = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("service error", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokePlayersService(t, m, http.MethodGet, service.PlayersRoute+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.getCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		player := arcade.Player{
			ID:          id,
			Name:        name,
			Description: description,
			HomeID:      homeID,
			LocationID:  locationID,
		}
		m := &mockPlayersStorage{t: t, playerID: id, player: player}

		w := invokePlayersService(t, m, http.MethodGet, service.PlayersRoute+"/"+id, nil)

		if !m.getCalled {
			t.Error("expected get to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body")
		}
		defer resp.Body.Close()

		var playerResp arcade.PlayerResponse
		err = json.Unmarshal(body, &playerResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := playerResp.Data
		if p.ID != id ||
			p.Name != name ||
			p.Description != description ||
			p.HomeID != homeID ||
			p.LocationID != locationID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestPlayersServiceCreate(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		homeID      = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID  = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodPost, service.PlayersRoute, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodPost, service.PlayersRoute, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","homeID": "` + homeID + `","locationID":"` + locationID + `"}`,
		)

		checkRespError(
			t, invokePlayersService(t, m, http.MethodPost, service.PlayersRoute, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.createCalled {
			t.Errorf("expected create to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := arcade.PlayerRequest{
			Name:        name,
			Description: description,
			HomeID:      homeID,
			LocationID:  locationID,
		}
		player := arcade.Player{
			ID:          id,
			Name:        name,
			Description: description,
			HomeID:      homeID,
			LocationID:  locationID,
			Created:     now,
			Updated:     now,
		}
		m := &mockPlayersStorage{t: t, req: req, player: player}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","homeID": "` + homeID + `","locationID":"` + locationID + `"}`,
		)

		w := invokePlayersService(t, m, http.MethodPost, service.PlayersRoute, body)

		if !m.createCalled {
			t.Errorf("expected create to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body")
		}
		defer resp.Body.Close()

		var playerResp arcade.PlayerResponse
		err = json.Unmarshal(b, &playerResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := playerResp.Data
		if p.ID != id ||
			p.Name != name ||
			p.Description != description ||
			p.HomeID != homeID ||
			p.LocationID != locationID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestPlayersServiceUpdate(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		homeID      = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID  = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodPut, service.PlayersRoute+"/"+id, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodPut, service.PlayersRoute+"/"+id, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","homeID": "` + homeID + `","locationID":"` + locationID + `"}`,
		)

		checkRespError(
			t, invokePlayersService(t, m, http.MethodPut, service.PlayersRoute+"/"+id, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.updateCalled {
			t.Errorf("expected update to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := arcade.PlayerRequest{
			Name:        name,
			Description: description,
			HomeID:      homeID,
			LocationID:  locationID,
		}
		player := arcade.Player{
			ID:          id,
			Name:        name,
			Description: description,
			HomeID:      homeID,
			LocationID:  locationID,
			Created:     now,
			Updated:     now,
		}
		m := &mockPlayersStorage{t: t, req: req, playerID: id, player: player}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","homeID": "` + homeID + `","locationID":"` + locationID + `"}`,
		)

		w := invokePlayersService(t, m, http.MethodPut, service.PlayersRoute+"/"+id, body)

		if !m.updateCalled {
			t.Errorf("expected update to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body")
		}
		defer resp.Body.Close()

		var playerResp arcade.PlayerResponse
		err = json.Unmarshal(b, &playerResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := playerResp.Data
		if p.ID != id ||
			p.Name != name ||
			p.Description != description ||
			p.HomeID != homeID ||
			p.LocationID != locationID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestPlayersServiceRemove(t *testing.T) {
	const (
		id = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
	)

	t.Run("service error", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokePlayersService(t, m, http.MethodDelete, service.PlayersRoute+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, playerID: id}

		w := invokePlayersService(t, m, http.MethodDelete, service.PlayersRoute+"/"+id, nil)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}

func invokePlayersService(t *testing.T, m *mockPlayersStorage, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	router := mux.NewRouter()
	s := service.PlayersService{Storage: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	return w
}

type (
	mockPlayersStorage struct {
		t   *testing.T
		err error

		playerID string
		req      arcade.PlayerRequest

		player  arcade.Player
		players []arcade.Player

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockPlayersStorage) List(context.Context, arcade.PlayersFilter) ([]arcade.Player, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.players, nil
}

func (m *mockPlayersStorage) Get(ctx context.Context, playerID string) (arcade.Player, error) {
	m.getCalled = true
	if m.err != nil {
		return arcade.Player{}, m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("get: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	return m.player, nil
}

func (m *mockPlayersStorage) Create(ctx context.Context, req arcade.PlayerRequest) (arcade.Player, error) {
	m.createCalled = true
	if m.err != nil {
		return arcade.Player{}, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected player request %+v, actual player requset %+v", m.req, req)
	}
	return m.player, nil
}

func (m *mockPlayersStorage) Update(ctx context.Context, playerID string, req arcade.PlayerRequest) (arcade.Player, error) {
	m.updateCalled = true
	if m.err != nil {
		return arcade.Player{}, m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("get: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected player request %+v, actual player requset %+v", m.req, req)
	}
	return m.player, nil
}

func (m *mockPlayersStorage) Remove(ctx context.Context, playerID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("remove: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	return nil
}
