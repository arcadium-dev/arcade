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

package rooms

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"arcadium.dev/arcade/internal/arcade"
	chttp "arcadium.dev/core/http"
)

func TestHandlerList(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		err := errors.New("unknown error")
		m := &mockService{t: t, err: err}

		checkRespError(
			t, invokeHandler(t, m, http.MethodGet, route, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.listCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		rooms := []arcade.Room{
			room{
				id:          "c39761fc-5096-4b1c-9d02-c75730b7b8bf",
				name:        "Drunen",
				description: "Son of Martin",
				owner:       "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				parent:      "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
			},
		}
		m := &mockService{t: t, rooms: rooms}

		w := invokeHandler(t, m, http.MethodGet, route, nil)

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

		var roomsResp roomsResponse
		err = json.Unmarshal(body, &roomsResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		if len(roomsResp.Data) != len(rooms) {
			t.Fatalf("Unexpected rooms response data length: %d", len(roomsResp.Data))
		}

		p := roomsResp.Data[0]
		if p.RoomID != rooms[0].ID() ||
			p.Name != rooms[0].Name() ||
			p.Description != rooms[0].Description() ||
			p.Owner != rooms[0].Owner() ||
			p.Parent != rooms[0].Parent() {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestHandlerGet(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		m := &mockService{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokeHandler(t, m, http.MethodGet, route+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.getCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		room := room{
			id:          id,
			name:        name,
			description: description,
			owner:       owner,
			parent:      parent,
		}
		m := &mockService{t: t, roomID: id, room: p}

		w := invokeHandler(t, m, http.MethodGet, route+"/"+id, nil)

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

		var roomResp roomResponse
		err = json.Unmarshal(body, &roomResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := roomResp.Data
		if p.RoomID != room.ID() ||
			p.Name != room.Name() ||
			p.Description != room.Description() ||
			p.Owner != room.Owner() ||
			p.Parent != room.Parent() {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestHandlerCreate(t *testing.T) {
	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokeHandler(t, nil, http.MethodPost, route, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokeHandler(t, nil, http.MethodPost, route, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockService{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","owner": "` + owner + `","parent":"` + parent + `"}`,
		)

		checkRespError(
			t, invokeHandler(t, m, http.MethodPost, route, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.createCalled {
			t.Errorf("expected create to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := roomRequest{
			Name:        name,
			Description: description,
			Owner:       owner,
			Parent:      parent,
		}
		room := room{
			id:          id,
			name:        name,
			description: description,
			owner:       owner,
			parent:      parent,
			created:     now,
			updated:     now,
		}
		m := &mockService{t: t, req: req, room: room}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","owner": "` + owner + `","parent":"` + parent + `"}`,
		)

		w := invokeHandler(t, m, http.MethodPost, route, body)

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

		var roomResp roomResponse
		err = json.Unmarshal(b, &roomResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := roomResp.Data
		if p.RoomID != room.ID() ||
			p.Name != room.Name() ||
			p.Description != room.Description() ||
			p.Owner != room.Owner() ||
			p.Parent != room.Parent() {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestHandlerUpdate(t *testing.T) {
	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokeHandler(t, nil, http.MethodPut, route+"/"+id, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokeHandler(t, nil, http.MethodPut, route+"/"+id, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockService{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","owner": "` + owner + `","parent":"` + parent + `"}`,
		)

		checkRespError(
			t, invokeHandler(t, m, http.MethodPut, route+"/"+id, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.updateCalled {
			t.Errorf("expected update to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := roomRequest{
			Name:        name,
			Description: description,
			Owner:       owner,
			Parent:      parent,
		}
		room := room{
			id:          id,
			name:        name,
			description: description,
			owner:       owner,
			parent:      parent,
			created:     now,
			updated:     now,
		}
		m := &mockService{t: t, req: req, roomID: id, room: room}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","owner": "` + owner + `","parent":"` + parent + `"}`,
		)

		w := invokeHandler(t, m, http.MethodPut, route+"/"+id, body)

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

		var roomResp roomResponse
		err = json.Unmarshal(b, &roomResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := roomResp.Data
		if p.RoomID != room.ID() ||
			p.Name != room.Name() ||
			p.Description != room.Description() ||
			p.Owner != room.Owner() ||
			p.Parent != room.Parent() {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestHandlerRemove(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		m := &mockService{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokeHandler(t, m, http.MethodDelete, route+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		m := &mockService{t: t, roomID: id}

		w := invokeHandler(t, m, http.MethodDelete, route+"/"+id, nil)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}

func invokeHandler(t *testing.T, m *mockService, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	router := mux.NewRouter()
	s := &Service{}
	s.h = handler{s: m}
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
