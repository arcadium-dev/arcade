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

package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"arcadium.dev/arcade"
	chttp "arcadium.dev/core/http"
)

func TestPlayersServiceName(t *testing.T) {
	s := PlayersService{}
	if s.Name() != "players" {
		t.Error("Unexpected service name")
	}
}

func TestPlayersServiceShutdown(t *testing.T) {
	// This is a placeholder for when we have a background monitor service running.
	s := PlayersService{}
	s.Shutdown()
}

func TestPlayersServiceList(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		err := errors.New("unknown error")
		m := &mockPlayersStorage{t: t, err: err}

		checkRespError(
			t, invokePlayersService(t, m, http.MethodGet, route, nil),
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

		w := invokePlayersService(t, m, http.MethodGet, route, nil)

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

/*
func TestPlayersServiceGet(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokePlayersService(t, m, http.MethodGet, route+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.getCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		player := player{
			id:          id,
			name:        name,
			description: description,
			home:        home,
			location:    location,
		}
		m := &mockPlayersStorage{t: t, playerID: id, player: p}

		w := invokePlayersService(t, m, http.MethodGet, route+"/"+id, nil)

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

		var playerResp playerResponse
		err = json.Unmarshal(body, &playerResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := playerResp.Data
		if p.PlayerID != player.ID() ||
			p.Name != player.Name() ||
			p.Description != player.Description() ||
			p.Home != player.Home() ||
			p.Location != player.Location() {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestPlayersServiceCreate(t *testing.T) {
	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodPost, route, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodPost, route, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","home": "` + home + `","location":"` + location + `"}`,
		)

		checkRespError(
			t, invokePlayersService(t, m, http.MethodPost, route, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.createCalled {
			t.Errorf("expected create to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := playerRequest{
			Name:        name,
			Description: description,
			Home:        home,
			Location:    location,
		}
		player := player{
			id:          id,
			name:        name,
			description: description,
			home:        home,
			location:    location,
			created:     now,
			updated:     now,
		}
		m := &mockPlayersStorage{t: t, req: req, player: player}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","home": "` + home + `","location":"` + location + `"}`,
		)

		w := invokePlayersService(t, m, http.MethodPost, route, body)

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

		var playerResp playerResponse
		err = json.Unmarshal(b, &playerResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := playerResp.Data
		if p.PlayerID != player.ID() ||
			p.Name != player.Name() ||
			p.Description != player.Description() ||
			p.Home != player.Home() ||
			p.Location != player.Location() {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestPlayersServiceUpdate(t *testing.T) {
	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodPut, route+"/"+id, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokePlayersService(t, nil, http.MethodPut, route+"/"+id, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","home": "` + home + `","location":"` + location + `"}`,
		)

		checkRespError(
			t, invokePlayersService(t, m, http.MethodPut, route+"/"+id, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.updateCalled {
			t.Errorf("expected update to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := playerRequest{
			Name:        name,
			Description: description,
			Home:        home,
			Location:    location,
		}
		player := player{
			id:          id,
			name:        name,
			description: description,
			home:        home,
			location:    location,
			created:     now,
			updated:     now,
		}
		m := &mockPlayersStorage{t: t, req: req, playerID: id, player: player}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","home": "` + home + `","location":"` + location + `"}`,
		)

		w := invokePlayersService(t, m, http.MethodPut, route+"/"+id, body)

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

		var playerResp playerResponse
		err = json.Unmarshal(b, &playerResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := playerResp.Data
		if p.PlayerID != player.ID() ||
			p.Name != player.Name() ||
			p.Description != player.Description() ||
			p.Home != player.Home() ||
			p.Location != player.Location() {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestPlayersServiceRemove(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokePlayersService(t, m, http.MethodDelete, route+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		m := &mockPlayersStorage{t: t, playerID: id}

		w := invokePlayersService(t, m, http.MethodDelete, route+"/"+id, nil)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}
*/

func invokePlayersService(t *testing.T, m *mockPlayersStorage, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	router := mux.NewRouter()
	s := &PlayersService{Storage: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	return w
}

func checkRespError(t *testing.T, w *httptest.ResponseRecorder, status int, errMsg string) {
	t.Helper()

	resp := w.Result()
	if resp.StatusCode != status {
		t.Errorf("Unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response body")
	}
	defer resp.Body.Close()

	var respErr struct {
		Error chttp.ResponseError `json:"error,omitempty"`
	}
	err = json.Unmarshal(body, &respErr)
	if err != nil {
		t.Errorf("Failed to json unmarshal error response: %s", err)
	}

	if !strings.Contains(respErr.Error.Detail, errMsg) {
		t.Errorf("\nExpected error detail: %s\nActual error detail:   %s", errMsg, respErr.Error.Detail)
	}
	if respErr.Error.Status != status {
		t.Errorf("Unexpected error status: %d", respErr.Error.Status)
	}
}
