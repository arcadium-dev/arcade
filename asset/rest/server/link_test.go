package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/errors"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
	"arcadium.dev/arcade/asset/rest/server"
)

func TestLinksList(t *testing.T) {
	route := server.V1LinkRoute
	id := uuid.New()

	t.Run("new filter failure", func(t *testing.T) {
		m := mockLinkStorage{}

		// ownerID failure
		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil, "ownerID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid ownerID query parameter: 'bad uuid'")

		// locationID failure
		w = invokeLinksEndpoint(t, m, http.MethodGet, route, nil, "locationID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid locationID query parameter: 'bad uuid'")

		// destinationID failure
		w = invokeLinksEndpoint(t, m, http.MethodGet, route, nil, "destinationID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid destinationID query parameter: 'bad uuid'")

		// offset failure
		w = invokeLinksEndpoint(t, m, http.MethodGet, route, nil, "offset", "bad offset")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid offset query parameter: 'bad offset'")

		// limit failure
		w = invokeLinksEndpoint(t, m, http.MethodGet, route, nil, "limit", "bad limit")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid limit query parameter: 'bad limit'")
	})

	t.Run("link manager list failure", func(t *testing.T) {
		m := mockLinkStorage{
			t: t,
			filter: asset.LinkFilter{
				LocationID: asset.RoomID(id),
				Offset:     10,
				Limit:      10,
			},
			listErr: fmt.Errorf("%w: list failure", errors.ErrNotFound),
		}

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "offset", "10", "limit", "10")
		assertRespError(t, w, http.StatusNotFound, "list failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			linkID        = asset.LinkID(uuid.New())
			ownerID       = asset.PlayerID(uuid.New())
			locationID    = asset.RoomID(uuid.New())
			destinationID = asset.RoomID(uuid.New())
			created       = arcade.Timestamp{Time: time.Now()}
			updated       = arcade.Timestamp{Time: time.Now()}
		)

		m := mockLinkStorage{
			t: t,
			filter: asset.LinkFilter{
				Offset: 25,
				Limit:  100,
			},
			list: []*asset.Link{
				{
					ID:            linkID,
					Name:          name,
					Description:   desc,
					OwnerID:       ownerID,
					LocationID:    locationID,
					DestinationID: destinationID,
					Created:       created,
					Updated:       updated,
				},
			},
		}

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil, "offset", "25", "limit", "100")

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressLinks rest.LinksResponse
		assert.Nil(t, json.Unmarshal(body, &egressLinks))

		assert.Compare(t, egressLinks, rest.LinksResponse{Links: []rest.Link{
			{
				ID:            linkID.String(),
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID.String(),
				LocationID:    locationID.String(),
				DestinationID: destinationID.String(),
				Created:       created,
				Updated:       updated,
			},
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestLinkGet(t *testing.T) {
	linkID := asset.LinkID(uuid.New())

	t.Run("linkID failure", func(t *testing.T) {
		m := mockLinkStorage{}

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid link id, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("link manager get failure", func(t *testing.T) {
		m := mockLinkStorage{
			t:      t,
			getID:  linkID,
			getErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			ownerID       = asset.PlayerID(uuid.New())
			locationID    = asset.RoomID(ownerID)
			destinationID = asset.RoomID(ownerID)
			created       = arcade.Timestamp{Time: time.Now()}
			updated       = arcade.Timestamp{Time: time.Now()}
		)

		m := mockLinkStorage{
			t:     t,
			getID: linkID,
			getLink: &asset.Link{
				ID:            linkID,
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
				Created:       created,
				Updated:       updated,
			},
		}

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressLink rest.LinkResponse
		assert.Nil(t, json.Unmarshal(body, &egressLink))

		assert.Compare(t, egressLink, rest.LinkResponse{Link: rest.Link{
			ID:            linkID.String(),
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locationID.String(),
			DestinationID: destinationID.String(),
			Created:       created,
			Updated:       updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestLinkCreate(t *testing.T) {
	route := server.V1LinkRoute

	t.Run("empty body", func(t *testing.T) {
		m := mockLinkStorage{}

		w := invokeLinksEndpoint(t, m, http.MethodPost, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeLinksEndpoint(t, m, http.MethodPost, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockLinkStorage{}

		w := invokeLinksEndpoint(t, m, http.MethodPost, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("create link req failure", func(t *testing.T) {
		m := mockLinkStorage{}

		tests := []struct {
			req    rest.LinkRequest
			status int
			errMsg string
		}{
			{
				req: rest.LinkRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link name",
			},
			{
				req: rest.LinkRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link name exceeds maximum length",
			},
			{
				req: rest.LinkRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link description",
			},
			{
				req: rest.LinkRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link description exceeds maximum length",
			},
			{
				req: rest.LinkRequest{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				req: rest.LinkRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID:  "bad location id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID: 'bad location id', invalid UUID length: 15",
			},
			{
				req: rest.LinkRequest{
					Name:          "name",
					Description:   "description",
					OwnerID:       uuid.New().String(),
					LocationID:    uuid.New().String(),
					DestinationID: "bad destination id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid destinationID: 'bad destination id', invalid UUID length: 18",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.LinkCreateRequest{LinkRequest: test.req})
			assert.Nil(t, err)

			w := invokeLinksEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("link manager create failure", func(t *testing.T) {
		var (
			ownerID = asset.PlayerID(uuid.New())
			locID   = asset.RoomID(uuid.New())
			destID  = asset.RoomID(uuid.New())
		)

		m := mockLinkStorage{
			t: t,
			create: asset.LinkCreate{
				LinkChange: asset.LinkChange{
					Name:          "name",
					Description:   "description",
					OwnerID:       ownerID,
					LocationID:    locID,
					DestinationID: destID,
				},
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		createReq := rest.LinkCreateRequest{
			LinkRequest: rest.LinkRequest{
				Name:          "name",
				Description:   "description",
				OwnerID:       ownerID.String(),
				LocationID:    locID.String(),
				DestinationID: destID.String(),
			},
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokeLinksEndpoint(t, m, http.MethodPost, route, body)
		assertRespError(t, w, http.StatusConflict, "conflict: create failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			linkID  = asset.LinkID(uuid.New())
			ownerID = asset.PlayerID(uuid.New())
			locID   = asset.RoomID(uuid.New())
			destID  = asset.RoomID(uuid.New())
			created = arcade.Timestamp{Time: time.Now()}
			updated = arcade.Timestamp{Time: time.Now()}
		)

		m := mockLinkStorage{
			t: t,
			create: asset.LinkCreate{
				LinkChange: asset.LinkChange{
					Name:          name,
					Description:   desc,
					OwnerID:       ownerID,
					LocationID:    locID,
					DestinationID: destID,
				},
			},
			createLink: &asset.Link{
				ID:            linkID,
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
				Created:       created,
				Updated:       updated,
			},
		}

		createReq := rest.LinkCreateRequest{
			LinkRequest: rest.LinkRequest{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID.String(),
				LocationID:    locID.String(),
				DestinationID: destID.String(),
			},
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokeLinksEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var linkResp rest.LinkResponse
		assert.Nil(t, json.Unmarshal(respBody, &linkResp))

		assert.Compare(t, linkResp, rest.LinkResponse{Link: rest.Link{
			ID:            linkID.String(),
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
			Created:       created,
			Updated:       updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestLinkUpdate(t *testing.T) {
	t.Run("linkID failure", func(t *testing.T) {
		m := mockLinkStorage{}

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid link id, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockLinkStorage{}

		linkID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeLinksEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockLinkStorage{}

		linkID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("update link req failure", func(t *testing.T) {
		m := mockLinkStorage{}

		tests := []struct {
			req    rest.LinkRequest
			status int
			errMsg string
		}{
			{
				req: rest.LinkRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link name",
			},
			{
				req: rest.LinkRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link name exceeds maximum length",
			},
			{
				req: rest.LinkRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link description",
			},
			{
				req: rest.LinkRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link description exceeds maximum length",
			},
			{
				req: rest.LinkRequest{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				req: rest.LinkRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID:  "bad location id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID: 'bad location id', invalid UUID length: 15",
			},
			{
				req: rest.LinkRequest{
					Name:          "name",
					Description:   "description",
					OwnerID:       uuid.New().String(),
					LocationID:    uuid.New().String(),
					DestinationID: "bad destination id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid destinationID: 'bad destination id', invalid UUID length: 18",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.LinkUpdateRequest{LinkRequest: test.req})
			assert.Nil(t, err)

			linkID := uuid.New()
			route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

			w := invokeLinksEndpoint(t, m, http.MethodPut, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("link manager update failure", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			linkID  = asset.LinkID(uuid.New())
			ownerID = asset.PlayerID(uuid.New())
			locID   = asset.RoomID(uuid.New())
			destID  = asset.RoomID(uuid.New())
		)

		m := mockLinkStorage{
			t:        t,
			updateID: linkID,
			update: asset.LinkUpdate{
				LinkChange: asset.LinkChange{
					Name:          name,
					Description:   desc,
					OwnerID:       ownerID,
					LocationID:    locID,
					DestinationID: destID,
				},
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		updateReq := rest.LinkUpdateRequest{
			LinkRequest: rest.LinkRequest{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID.String(),
				LocationID:    locID.String(),
				DestinationID: destID.String(),
			},
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			linkID  = asset.LinkID(uuid.New())
			ownerID = asset.PlayerID(uuid.New())
			locID   = asset.RoomID(uuid.New())
			destID  = asset.RoomID(uuid.New())
			created = arcade.Timestamp{Time: time.Now()}
			updated = arcade.Timestamp{Time: time.Now()}
		)

		m := mockLinkStorage{
			t:        t,
			updateID: linkID,
			update: asset.LinkUpdate{
				LinkChange: asset.LinkChange{
					Name:          name,
					Description:   desc,
					OwnerID:       ownerID,
					LocationID:    locID,
					DestinationID: destID,
				},
			},
			updateLink: &asset.Link{
				ID:            linkID,
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
				Created:       created,
				Updated:       updated,
			},
		}

		updateReq := rest.LinkUpdateRequest{
			LinkRequest: rest.LinkRequest{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID.String(),
				LocationID:    locID.String(),
				DestinationID: destID.String(),
			},
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var linkResp rest.LinkResponse
		assert.Nil(t, json.Unmarshal(respBody, &linkResp))

		assert.Compare(t, linkResp, rest.LinkResponse{Link: rest.Link{
			ID:            linkID.String(),
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
			Created:       created,
			Updated:       updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestLinkRemove(t *testing.T) {
	linkID := asset.LinkID(uuid.New())

	t.Run("linkID failure", func(t *testing.T) {
		m := mockLinkStorage{}

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid link id, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("link manager remove failure", func(t *testing.T) {
		m := mockLinkStorage{
			t:            t,
			removeLinkID: linkID,
			removeErr:    fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockLinkStorage{
			t:            t,
			removeLinkID: linkID,
		}

		route := fmt.Sprintf("%s/%s", server.V1LinkRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// helper

func invokeLinksEndpoint(t *testing.T, m mockLinkStorage, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := server.LinksService{Storage: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, b)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	for i := 0; i < len(query); i += 2 {
		q.Add(query[i], query[i+1])
	}
	r.URL.RawQuery = q.Encode()

	router.ServeHTTP(w, r)

	return w
}

// mockLinkStorage

type (
	mockLinkStorage struct {
		t *testing.T

		filter  asset.LinkFilter
		list    []*asset.Link
		listErr error

		getID   asset.LinkID
		getLink *asset.Link
		getErr  error

		create     asset.LinkCreate
		createLink *asset.Link
		createErr  error

		updateID   asset.LinkID
		update     asset.LinkUpdate
		updateLink *asset.Link
		updateErr  error

		removeLinkID asset.LinkID
		removeErr    error
	}
)

func (m mockLinkStorage) List(ctx context.Context, filter asset.LinkFilter) ([]*asset.Link, error) {
	assert.Compare(m.t, filter, m.filter)
	return m.list, m.listErr
}

func (m mockLinkStorage) Get(ctx context.Context, id asset.LinkID) (*asset.Link, error) {
	assert.Compare(m.t, id, m.getID)
	return m.getLink, m.getErr
}

func (m mockLinkStorage) Create(ctx context.Context, create asset.LinkCreate) (*asset.Link, error) {
	assert.Compare(m.t, create, m.create)
	return m.createLink, m.createErr
}

func (m mockLinkStorage) Update(ctx context.Context, id asset.LinkID, update asset.LinkUpdate) (*asset.Link, error) {
	assert.Compare(m.t, id, m.updateID)
	assert.Compare(m.t, update, m.update)
	return m.updateLink, m.updateErr
}

func (m mockLinkStorage) Remove(ctx context.Context, id asset.LinkID) error {
	assert.Compare(m.t, id, m.removeLinkID)
	return m.removeErr
}
