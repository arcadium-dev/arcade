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

func TestLinksName(t *testing.T) {
	s := service.Links{}
	if s.Name() != "links" {
		t.Error("Unexpected service name")
	}
}

func TestLinksList(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		err := errors.New("unknown error")
		m := &mockLinksStorage{t: t, err: err}

		checkRespError(
			t, invokeLinks(t, m, http.MethodGet, service.LinksRoute, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.listCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		links := []arcade.Link{
			{
				ID:            "c39761fc-5096-4b1c-9d02-c75730b7b8bf",
				Name:          "Drunen",
				Description:   "Son of Martin",
				OwnerID:       "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				LocationID:    "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
				DestinationID: "2564cd4e-ae30-42a9-aaea-a1203ef0414b",
			},
		}
		m := &mockLinksStorage{t: t, links: links}

		w := invokeLinks(t, m, http.MethodGet, service.LinksRoute, nil)

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

		var linksResp arcade.LinksResponse
		err = json.Unmarshal(body, &linksResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		if len(linksResp.Data) != len(links) {
			t.Fatalf("Unexpected links response data length: %d", len(linksResp.Data))
		}

		link := linksResp.Data[0]
		if link.ID != links[0].ID ||
			link.Name != links[0].Name ||
			link.Description != links[0].Description ||
			link.OwnerID != links[0].OwnerID ||
			link.LocationID != links[0].LocationID ||
			link.DestinationID != links[0].DestinationID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestLinksGet(t *testing.T) {
	const (
		id            = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name          = "Drunen"
		description   = "Son of Martin"
		ownerID       = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID    = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		destinationID = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("service error", func(t *testing.T) {
		m := &mockLinksStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokeLinks(t, m, http.MethodGet, service.LinksRoute+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.getCalled {
			t.Error("expected list to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		link := arcade.Link{
			ID:            id,
			Name:          name,
			Description:   description,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
		}
		m := &mockLinksStorage{t: t, linkID: id, link: link}

		w := invokeLinks(t, m, http.MethodGet, service.LinksRoute+"/"+id, nil)

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

		var linkResp arcade.LinkResponse
		err = json.Unmarshal(body, &linkResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := linkResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.LocationID != locationID ||
			r.DestinationID != destinationID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestLinksCreate(t *testing.T) {
	const (
		id            = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name          = "Drunen"
		description   = "Son of Martin"
		ownerID       = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID    = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		destinationID = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokeLinks(t, nil, http.MethodPost, service.LinksRoute, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokeLinks(t, nil, http.MethodPost, service.LinksRoute, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockLinksStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","locationID":"` + locationID + `","destinationID":"` + destinationID + `"}`,
		)

		checkRespError(
			t, invokeLinks(t, m, http.MethodPost, service.LinksRoute, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.createCalled {
			t.Errorf("expected create to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := arcade.LinkRequest{
			Name:          name,
			Description:   description,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
		}
		link := arcade.Link{
			ID:            id,
			Name:          name,
			Description:   description,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
			Created:       now,
			Updated:       now,
		}
		m := &mockLinksStorage{t: t, req: req, link: link}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","locationID":"` + locationID + `","destinationID":"` + destinationID + `"}`,
		)

		w := invokeLinks(t, m, http.MethodPost, service.LinksRoute, body)

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

		var linkResp arcade.LinkResponse
		err = json.Unmarshal(b, &linkResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := linkResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.LocationID != locationID ||
			r.DestinationID != destinationID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestLinksUpdate(t *testing.T) {
	const (
		id            = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
		name          = "Drunen"
		description   = "Son of Martin"
		ownerID       = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		locationID    = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
		destinationID = "2564cd4e-ae30-42a9-aaea-a1203ef0414b"
	)

	t.Run("missing body", func(t *testing.T) {
		checkRespError(
			t, invokeLinks(t, nil, http.MethodPut, service.LinksRoute+"/"+id, nil),
			http.StatusBadRequest, "invalid argument: invalid json: a json encoded body is required",
		)
	})

	t.Run("invalid json", func(t *testing.T) {
		checkRespError(
			t, invokeLinks(t, nil, http.MethodPut, service.LinksRoute+"/"+id, bytes.NewBufferString(`invalid json`)),
			http.StatusBadRequest, "invalid argument: invalid body: ",
		)
	})

	t.Run("service error", func(t *testing.T) {
		m := &mockLinksStorage{t: t, err: errors.New("unknown error")}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","locationID":"` + locationID + `","destinationID":"` + destinationID + `"}`,
		)

		checkRespError(
			t, invokeLinks(t, m, http.MethodPut, service.LinksRoute+"/"+id, body),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.updateCalled {
			t.Errorf("expected update to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		req := arcade.LinkRequest{
			Name:          name,
			Description:   description,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
		}
		link := arcade.Link{
			ID:            id,
			Name:          name,
			Description:   description,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
			Created:       now,
			Updated:       now,
		}
		m := &mockLinksStorage{t: t, req: req, linkID: id, link: link}
		body := bytes.NewBufferString(
			`{"name":"` + name + `","description":"` + description + `","ownerID": "` + ownerID + `","locationID":"` + locationID + `","destinationID":"` + destinationID + `"}`,
		)

		w := invokeLinks(t, m, http.MethodPut, service.LinksRoute+"/"+id, body)

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

		var linkResp arcade.LinkResponse
		err = json.Unmarshal(b, &linkResp)
		if err != nil {
			t.Errorf("Failed to json unmarshal response: %s", err)
		}

		r := linkResp.Data
		if r.ID != id ||
			r.Name != name ||
			r.Description != description ||
			r.OwnerID != ownerID ||
			r.LocationID != locationID ||
			r.DestinationID != destinationID {
			t.Errorf("Unexpected response data")
		}
	})
}

func TestLinksRemove(t *testing.T) {
	const (
		id = "c39761fc-5096-4b1c-9d02-c75730b7b8bf"
	)

	t.Run("service error", func(t *testing.T) {
		m := &mockLinksStorage{t: t, err: errors.New("unknown error")}

		checkRespError(
			t, invokeLinks(t, m, http.MethodDelete, service.LinksRoute+"/"+id, nil),
			http.StatusInternalServerError, "unknown error",
		)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		m := &mockLinksStorage{t: t, linkID: id}

		w := invokeLinks(t, m, http.MethodDelete, service.LinksRoute+"/"+id, nil)

		if !m.removeCalled {
			t.Error("expected remove to be called")
		}
		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}

func invokeLinks(t *testing.T, m *mockLinksStorage, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	router := mux.NewRouter()
	s := service.Links{Storage: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	return w
}

type (
	mockLinksStorage struct {
		t   *testing.T
		err error

		linkID string
		req    arcade.LinkRequest

		link  arcade.Link
		links []arcade.Link

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockLinksStorage) List(context.Context, arcade.LinksFilter) ([]arcade.Link, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.links, nil
}

func (m *mockLinksStorage) Get(ctx context.Context, linkID string) (arcade.Link, error) {
	m.getCalled = true
	if m.err != nil {
		return arcade.Link{}, m.err
	}
	if m.linkID != linkID {
		m.t.Fatalf("get: expected linkID %s, actual linkID %s", m.linkID, linkID)
	}
	return m.link, nil
}

func (m *mockLinksStorage) Create(ctx context.Context, req arcade.LinkRequest) (arcade.Link, error) {
	m.createCalled = true
	if m.err != nil {
		return arcade.Link{}, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected link request %+v, actual link requset %+v", m.req, req)
	}
	return m.link, nil
}

func (m *mockLinksStorage) Update(ctx context.Context, linkID string, req arcade.LinkRequest) (arcade.Link, error) {
	m.updateCalled = true
	if m.err != nil {
		return arcade.Link{}, m.err
	}
	if m.linkID != linkID {
		m.t.Fatalf("get: expected linkID %s, actual linkID %s", m.linkID, linkID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected link request %+v, actual link requset %+v", m.req, req)
	}
	return m.link, nil
}

func (m *mockLinksStorage) Remove(ctx context.Context, linkID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.linkID != linkID {
		m.t.Fatalf("remove: expected linkID %s, actual linkID %s", m.linkID, linkID)
	}
	return nil
}
