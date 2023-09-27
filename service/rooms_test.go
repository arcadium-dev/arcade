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

func TestRoomsServiceName(t *testing.T) {
	s := service.RoomsService{}
	if s.Name() != "rooms" {
		t.Error("Unexpected service name")
	}
}

func TestRoomsServiceList(t *testing.T) {
	t.Run("filter error", func(t *testing.T) {
		route := fmt.Sprintf("%s?ownerID=42", service.RoomsRoute)
		checkRespError(
			t, invokeRoomsService(t, nil, http.MethodGet, route, nil),
			http.StatusBadRequest,
			"invalid argument: invalid ownerID query parameter: '42'",
		)
	})

	t.Run("service error", func(t *testing.T) {
		err := errors.New("unknown error")
		m := &mockRoomsStorage{t: t, err: err}

		checkRespError(
			t, invokeRoomsService(t, m, http.MethodGet, service.RoomsRoute, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.listCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		rooms := []arcade.Room{
			{
				ID:          "c39761fc-5096-4b1c-9d02-c75730b7b8bf",
				Name:        "Drunen",
				Description: "Son of Martin",
				OwnerID:     "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				ParentID:    "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
			},
		}
		m := &mockRoomsStorage{t: t, rooms: rooms}

		w := invokeRoomsService(t, m, http.MethodGet, service.RoomsRoute, nil)

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

		var roomsResp arcade.RoomsResponse
		err = json.Unmarshal(body, &roomsResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		if len(roomsResp.Data) != len(rooms) {
			t.Fatalf("Unexpected rooms response data length: %d", len(roomsResp.Data))
		}

		room := roomsResp.Data[0]
		if room.ID != rooms[0].ID ||
			room.Name != rooms[0].Name ||
			room.Description != rooms[0].Description ||
			room.OwnerID != rooms[0].OwnerID ||
			room.ParentID != rooms[0].ParentID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestRoomsServiceGet(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		ownerID     = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		parentID    = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("service error", func(t *testing.T) {
		m := &mockRoomsStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokeRoomsService(t, m, http.MethodGet, service.RoomsRoute+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.getCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		room := arcade.Room{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			ParentID:    parentID,
		}
		m := &mockRoomsStorage{t: t, roomID: id, room: room}

		w := invokeRoomsService(t, m, http.MethodGet, service.RoomsRoute+"/"+id, nil)

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

		var roomResp arcade.RoomResponse
		err = json.Unmarshal(body, &roomResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := roomResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.ParentID != parentID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestRoomsServiceCreate(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		ownerID     = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		parentID    = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokeRoomsService(t, nil, http.MethodPost, service.RoomsRoute, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokeRoomsService(t, nil, http.MethodPost, service.RoomsRoute, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockRoomsStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","parentID":"` + parentID + `"}`,
		)

		checkRespError(
			t, invokeRoomsService(t, m, http.MethodPost, service.RoomsRoute, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.createCalled {
			t.Errorf("expected create to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := arcade.RoomRequest{
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			ParentID:    parentID,
		}
		room := arcade.Room{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			ParentID:    parentID,
			Created:     now,
			Updated:     now,
		}
		m := &mockRoomsStorage{t: t, req: req, room: room}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","parentID":"` + parentID + `"}`,
		)

		w := invokeRoomsService(t, m, http.MethodPost, service.RoomsRoute, body)

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

		var roomResp arcade.RoomResponse
		err = json.Unmarshal(b, &roomResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := roomResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.ParentID != parentID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestRoomsServiceUpdate(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		ownerID     = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		parentID    = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokeRoomsService(t, nil, http.MethodPut, service.RoomsRoute+"/"+id, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokeRoomsService(t, nil, http.MethodPut, service.RoomsRoute+"/"+id, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockRoomsStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","parentID":"` + parentID + `"}`,
		)

		checkRespError(
			t, invokeRoomsService(t, m, http.MethodPut, service.RoomsRoute+"/"+id, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.updateCalled {
			t.Errorf("expected update to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := arcade.RoomRequest{
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			ParentID:    parentID,
		}
		room := arcade.Room{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			ParentID:    parentID,
			Created:     now,
			Updated:     now,
		}
		m := &mockRoomsStorage{t: t, req: req, roomID: id, room: room}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","parentID":"` + parentID + `"}`,
		)

		w := invokeRoomsService(t, m, http.MethodPut, service.RoomsRoute+"/"+id, body)

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

		var roomResp arcade.RoomResponse
		err = json.Unmarshal(b, &roomResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := roomResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.ParentID != parentID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestRoomsServiceRemove(t *testing.T) {
	const (
		id = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
	)

	t.Run("service error", func(t *testing.T) {
		m := &mockRoomsStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokeRoomsService(t, m, http.MethodDelete, service.RoomsRoute+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		m := &mockRoomsStorage{t: t, roomID: id}

		w := invokeRoomsService(t, m, http.MethodDelete, service.RoomsRoute+"/"+id, nil)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}

func invokeRoomsService(t *testing.T, m *mockRoomsStorage, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	router := mux.NewRouter()
	s := service.RoomsService{Storage: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	return w
}

type (
	mockRoomsStorage struct {
		t   *testing.T
		err error

		roomID string
		req    arcade.RoomRequest

		room  arcade.Room
		rooms []arcade.Room

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockRoomsStorage) List(context.Context, arcade.RoomsFilter) ([]arcade.Room, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.rooms, nil
}

func (m *mockRoomsStorage) Get(ctx context.Context, roomID string) (arcade.Room, error) {
	m.getCalled = true
	if m.err != nil {
		return arcade.Room{}, m.err
	}
	if m.roomID != roomID {
		m.t.Fatalf("get: expected roomID %s, actual roomID %s", m.roomID, roomID)
	}
	return m.room, nil
}

func (m *mockRoomsStorage) Create(ctx context.Context, req arcade.RoomRequest) (arcade.Room, error) {
	m.createCalled = true
	if m.err != nil {
		return arcade.Room{}, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected room request %+v, actual room requset %+v", m.req, req)
	}
	return m.room, nil
}

func (m *mockRoomsStorage) Update(ctx context.Context, roomID string, req arcade.RoomRequest) (arcade.Room, error) {
	m.updateCalled = true
	if m.err != nil {
		return arcade.Room{}, m.err
	}
	if m.roomID != roomID {
		m.t.Fatalf("get: expected roomID %s, actual roomID %s", m.roomID, roomID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected room request %+v, actual room requset %+v", m.req, req)
	}
	return m.room, nil
}

func (m *mockRoomsStorage) Remove(ctx context.Context, roomID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.roomID != roomID {
		m.t.Fatalf("remove: expected roomID %s, actual roomID %s", m.roomID, roomID)
	}
	return nil
}
