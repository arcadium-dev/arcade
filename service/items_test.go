package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/service"
)

func TestItemsName(t *testing.T) {
	s := service.Items{}
	if s.Name() != "items" {
		t.Error("Unexpected service name")
	}
}

func TestItemsList(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		err := errors.New("unknown error")
		m := &mockItemsStorage{t: t, err: err}

		checkRespError(
			t, invokeItems(t, m, http.MethodGet, service.ItemsRoute, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.listCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		items := []arcade.Item{
			{
				ID:          "c39761fc-5096-4b1c-9d02-c75730b7b8bf",
				Name:        "Drunen",
				Description: "Son of Martin",
				OwnerID:     "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				LocationID:  "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				InventoryID: "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
			},
		}
		m := &mockItemsStorage{t: t, items: items}

		w := invokeItems(t, m, http.MethodGet, service.ItemsRoute, nil)

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

		var itemsResp arcade.ItemsResponse
		err = json.Unmarshal(body, &itemsResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		if len(itemsResp.Data) != len(items) {
			t.Fatalf("Unexpected items response data length: %d", len(itemsResp.Data))
		}

		item := itemsResp.Data[0]
		if item.ID != items[0].ID ||
			item.Name != items[0].Name ||
			item.Description != items[0].Description ||
			item.OwnerID != items[0].OwnerID ||
			item.LocationID != items[0].LocationID ||
			item.InventoryID != items[0].InventoryID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestItemsGet(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		ownerID     = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID  = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		inventoryID = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("service error", func(t *testing.T) {
		m := &mockItemsStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokeItems(t, m, http.MethodGet, service.ItemsRoute+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.getCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		item := arcade.Item{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			LocationID:  locationID,
			InventoryID: inventoryID,
		}
		m := &mockItemsStorage{t: t, itemID: id, item: item}

		w := invokeItems(t, m, http.MethodGet, service.ItemsRoute+"/"+id, nil)

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

		var itemResp arcade.ItemResponse
		err = json.Unmarshal(body, &itemResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := itemResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.LocationID != locationID ||
			r.InventoryID != inventoryID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestItemsCreate(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		ownerID     = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID  = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		inventoryID = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokeItems(t, nil, http.MethodPost, service.ItemsRoute, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokeItems(t, nil, http.MethodPost, service.ItemsRoute, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockItemsStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","locationID":"` + locationID + `","inventoryID":"` + inventoryID + `"}`,
		)

		checkRespError(
			t, invokeItems(t, m, http.MethodPost, service.ItemsRoute, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.createCalled {
			t.Errorf("expected create to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := arcade.ItemRequest{
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			LocationID:  locationID,
			InventoryID: inventoryID,
		}
		item := arcade.Item{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			LocationID:  locationID,
			InventoryID: inventoryID,
			Created:     now,
			Updated:     now,
		}
		m := &mockItemsStorage{t: t, req: req, item: item}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","locationID":"` + locationID + `","inventoryID":"` + inventoryID + `"}`,
		)

		w := invokeItems(t, m, http.MethodPost, service.ItemsRoute, body)

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

		var itemResp arcade.ItemResponse
		err = json.Unmarshal(b, &itemResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := itemResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.LocationID != locationID ||
			r.InventoryID != inventoryID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestItemsUpdate(t *testing.T) {
	const (
		id          = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name        = "Drunen"
		description = "Son of Martin"
		ownerID     = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID  = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		inventoryID = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokeItems(t, nil, http.MethodPut, service.ItemsRoute+"/"+id, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokeItems(t, nil, http.MethodPut, service.ItemsRoute+"/"+id, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockItemsStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","locationID":"` + locationID + `","inventoryID":"` + inventoryID + `"}`,
		)

		checkRespError(
			t, invokeItems(t, m, http.MethodPut, service.ItemsRoute+"/"+id, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.updateCalled {
			t.Errorf("expected update to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := arcade.ItemRequest{
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			LocationID:  locationID,
			InventoryID: inventoryID,
		}
		item := arcade.Item{
			ID:          id,
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
			LocationID:  locationID,
			InventoryID: inventoryID,
			Created:     now,
			Updated:     now,
		}
		m := &mockItemsStorage{t: t, req: req, itemID: id, item: item}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","locationID":"` + locationID + `","inventoryID":"` + inventoryID + `"}`,
		)

		w := invokeItems(t, m, http.MethodPut, service.ItemsRoute+"/"+id, body)

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

		var itemResp arcade.ItemResponse
		err = json.Unmarshal(b, &itemResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := itemResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.LocationID != locationID ||
			r.InventoryID != inventoryID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestItemsRemove(t *testing.T) {
	const (
		id = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
	)

	t.Run("service error", func(t *testing.T) {
		m := &mockItemsStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokeItems(t, m, http.MethodDelete, service.ItemsRoute+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		m := &mockItemsStorage{t: t, itemID: id}

		w := invokeItems(t, m, http.MethodDelete, service.ItemsRoute+"/"+id, nil)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}

func invokeItems(t *testing.T, m *mockItemsStorage, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	router := mux.NewRouter()
	s := service.Items{Storage: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	return w
}

type (
	mockItemsStorage struct {
		t   *testing.T
		err error

		itemID string
		req    arcade.ItemRequest

		item  arcade.Item
		items []arcade.Item

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockItemsStorage) List(context.Context, arcade.ItemsFilter) ([]arcade.Item, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func (m *mockItemsStorage) Get(ctx context.Context, itemID string) (arcade.Item, error) {
	m.getCalled = true
	if m.err != nil {
		return arcade.Item{}, m.err
	}
	if m.itemID != itemID {
		m.t.Fatalf("get: expected itemID %s, actual itemID %s", m.itemID, itemID)
	}
	return m.item, nil
}

func (m *mockItemsStorage) Create(ctx context.Context, req arcade.ItemRequest) (arcade.Item, error) {
	m.createCalled = true
	if m.err != nil {
		return arcade.Item{}, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected item request %+v, actual item requset %+v", m.req, req)
	}
	return m.item, nil
}

func (m *mockItemsStorage) Update(ctx context.Context, itemID string, req arcade.ItemRequest) (arcade.Item, error) {
	m.updateCalled = true
	if m.err != nil {
		return arcade.Item{}, m.err
	}
	if m.itemID != itemID {
		m.t.Fatalf("get: expected itemID %s, actual itemID %s", m.itemID, itemID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected item request %+v, actual item requset %+v", m.req, req)
	}
	return m.item, nil
}

func (m *mockItemsStorage) Remove(ctx context.Context, itemID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.itemID != itemID {
		m.t.Fatalf("remove: expected itemID %s, actual itemID %s", m.itemID, itemID)
	}
	return nil
}
