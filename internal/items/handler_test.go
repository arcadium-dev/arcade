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

package items

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
		items := []arcade.Item{
			item{
				id:          "c39761fc-5096-4b1c-9d02-c75730b7b8bf",
				name:        "Drunen",
				description: "Son of Martin",
				owner:       "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				location:    "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				inventory:   "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
			},
		}
		m := &mockService{t: t, items: items}

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

		var itemsResp itemsResponse
		err = json.Unmarshal(body, &itemsResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		if len(itemsResp.Data) != len(items) {
			t.Fatalf("Unexpected items response data length: %d", len(itemsResp.Data))
		}

		p := itemsResp.Data[0]
		if p.ItemID != items[0].ID() ||
			p.Name != items[0].Name() ||
			p.Description != items[0].Description() ||
			p.Owner != items[0].Owner() ||
			p.Location != items[0].Location() ||
			p.Inventory != items[0].Inventory() {
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
		item := item{
			id:          id,
			name:        name,
			description: description,
			owner:       owner,
			location:    location,
			inventory:   inventory,
		}
		m := &mockService{t: t, itemID: id, item: p}

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

		var itemResp itemResponse
		err = json.Unmarshal(body, &itemResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := itemResp.Data
		if p.ItemID != item.ID() ||
			p.Name != item.Name() ||
			p.Description != item.Description() ||
			p.Owner != item.Owner() ||
			p.Location != item.Location() ||
			p.Inventory != item.Inventory() {
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
			`{"name":"` + name + `","description":"` + description +
				`","owner": "` + owner + `","location":"` + location +
				`","inventory": "` + inventory + `"}`,
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
		req := itemRequest{
			Name:        name,
			Description: description,
			Owner:       owner,
			Location:    location,
			Inventory:   inventory,
		}
		item := item{
			id:          id,
			name:        name,
			description: description,
			owner:       owner,
			location:    location,
			inventory:   inventory,
			created:     now,
			updated:     now,
		}
		m := &mockService{t: t, req: req, item: item}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description +
				`","owner": "` + owner + `","location":"` + location +
				`","inventory": "` + inventory + `"}`,
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

		var itemResp itemResponse
		err = json.Unmarshal(b, &itemResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := itemResp.Data
		if p.ItemID != item.ID() ||
			p.Name != item.Name() ||
			p.Description != item.Description() ||
			p.Owner != item.Owner() ||
			p.Location != item.Location() ||
			p.Inventory != item.Inventory() {
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
			`{"name":"` + name + `","description":"` + description +
				`","owner": "` + owner + `","location":"` + location +
				`","inventory": "` + inventory + `"}`,
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
		req := itemRequest{
			Name:        name,
			Description: description,
			Owner:       owner,
			Location:    location,
			Inventory:   inventory,
		}
		item := item{
			id:          id,
			name:        name,
			description: description,
			owner:       owner,
			location:    location,
			inventory:   inventory,
			created:     now,
			updated:     now,
		}
		m := &mockService{t: t, req: req, itemID: id, item: item}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description +
				`","owner": "` + owner + `","location":"` + location +
				`","inventory": "` + inventory + `"}`,
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

		var itemResp itemResponse
		err = json.Unmarshal(b, &itemResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		p := itemResp.Data
		if p.ItemID != item.ID() ||
			p.Name != item.Name() ||
			p.Description != item.Description() ||
			p.Owner != item.Owner() ||
			p.Location != item.Location() ||
			p.Inventory != item.Inventory() {
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
		m := &mockService{t: t, itemID: id}

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
