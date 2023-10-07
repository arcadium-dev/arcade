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

	"arcadium.dev/arcade/assets"
	"arcadium.dev/arcade/assets/network"
	"arcadium.dev/arcade/assets/network/server"
)

func TestLinksList(t *testing.T) {
	route := server.V1LinksRoute
	id := uuid.New()

	t.Run("new filter failure", func(t *testing.T) {
		m := mockLinkManager{}

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
		m := mockLinkManager{
			t: t,
			filter: assets.LinksFilter{
				LocationID: assets.RoomID(id),
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
			linkID        = assets.LinkID(uuid.New())
			ownerID       = assets.PlayerID(uuid.New())
			locationID    = assets.RoomID(uuid.New())
			destinationID = assets.RoomID(uuid.New())
			created       = assets.Timestamp{Time: time.Now()}
			updated       = assets.Timestamp{Time: time.Now()}
		)

		m := mockLinkManager{
			t: t,
			filter: assets.LinksFilter{
				Offset: 25,
				Limit:  100,
			},
			list: []*assets.Link{
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

		var egressLinks network.LinksResponse
		assert.Nil(t, json.Unmarshal(body, &egressLinks))

		assert.Compare(t, egressLinks, network.LinksResponse{Links: []network.Link{
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
	linkID := assets.LinkID(uuid.New())

	t.Run("linkID failure", func(t *testing.T) {
		m := mockLinkManager{}

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid link id, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("link manager get failure", func(t *testing.T) {
		m := mockLinkManager{
			t:      t,
			getID:  linkID,
			getErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			ownerID       = assets.PlayerID(uuid.New())
			locationID    = assets.RoomID(ownerID)
			destinationID = assets.RoomID(ownerID)
			created       = assets.Timestamp{Time: time.Now()}
			updated       = assets.Timestamp{Time: time.Now()}
		)

		m := mockLinkManager{
			t:     t,
			getID: linkID,
			getLink: &assets.Link{
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

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressLink network.LinkResponse
		assert.Nil(t, json.Unmarshal(body, &egressLink))

		assert.Compare(t, egressLink, network.LinkResponse{Link: network.Link{
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
	route := server.V1LinksRoute

	t.Run("empty body", func(t *testing.T) {
		m := mockLinkManager{}

		w := invokeLinksEndpoint(t, m, http.MethodPost, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeLinksEndpoint(t, m, http.MethodPost, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockLinkManager{}

		w := invokeLinksEndpoint(t, m, http.MethodPost, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("create link req failure", func(t *testing.T) {
		m := mockLinkManager{}

		tests := []struct {
			req    network.LinkCreateRequest
			status int
			errMsg string
		}{
			{
				req: network.LinkCreateRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link name",
			},
			{
				req: network.LinkCreateRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link name exceeds maximum length",
			},
			{
				req: network.LinkCreateRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link description",
			},
			{
				req: network.LinkCreateRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link description exceeds maximum length",
			},
			{
				req: network.LinkCreateRequest{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				req: network.LinkCreateRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID:  "bad location id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID: 'bad location id', invalid UUID length: 15",
			},
			{
				req: network.LinkCreateRequest{
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
			body, err := json.Marshal(test.req)
			assert.Nil(t, err)

			w := invokeLinksEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("link manager create failure", func(t *testing.T) {
		var (
			ownerID = assets.PlayerID(uuid.New())
			locID   = assets.RoomID(uuid.New())
			destID  = assets.RoomID(uuid.New())
		)

		m := mockLinkManager{
			t: t,
			createReq: assets.LinkCreateRequest{
				Name:          "name",
				Description:   "description",
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		createReq := network.LinkCreateRequest{
			Name:          "name",
			Description:   "description",
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
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
			linkID  = assets.LinkID(uuid.New())
			ownerID = assets.PlayerID(uuid.New())
			locID   = assets.RoomID(uuid.New())
			destID  = assets.RoomID(uuid.New())
			created = assets.Timestamp{Time: time.Now()}
			updated = assets.Timestamp{Time: time.Now()}
		)

		m := mockLinkManager{
			t: t,
			createReq: assets.LinkCreateRequest{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
			},
			createLink: &assets.Link{
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

		createReq := network.LinkCreateRequest{
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokeLinksEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var linkResp network.LinkResponse
		assert.Nil(t, json.Unmarshal(respBody, &linkResp))

		assert.Compare(t, linkResp, network.LinkResponse{Link: network.Link{
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
		m := mockLinkManager{}

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid link id, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockLinkManager{}

		linkID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeLinksEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockLinkManager{}

		linkID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("update link req failure", func(t *testing.T) {
		m := mockLinkManager{}

		tests := []struct {
			req    network.LinkUpdateRequest
			status int
			errMsg string
		}{
			{
				req: network.LinkUpdateRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link name",
			},
			{
				req: network.LinkUpdateRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link name exceeds maximum length",
			},
			{
				req: network.LinkUpdateRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link description",
			},
			{
				req: network.LinkUpdateRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link description exceeds maximum length",
			},
			{
				req: network.LinkUpdateRequest{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				req: network.LinkUpdateRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID:  "bad location id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID: 'bad location id', invalid UUID length: 15",
			},
			{
				req: network.LinkUpdateRequest{
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
			body, err := json.Marshal(test.req)
			assert.Nil(t, err)

			linkID := uuid.New()
			route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

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
			linkID  = assets.LinkID(uuid.New())
			ownerID = assets.PlayerID(uuid.New())
			locID   = assets.RoomID(uuid.New())
			destID  = assets.RoomID(uuid.New())
		)

		m := mockLinkManager{
			t:        t,
			updateID: linkID,
			updateReq: assets.LinkUpdateRequest{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		updateReq := network.LinkUpdateRequest{
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			linkID  = assets.LinkID(uuid.New())
			ownerID = assets.PlayerID(uuid.New())
			locID   = assets.RoomID(uuid.New())
			destID  = assets.RoomID(uuid.New())
			created = assets.Timestamp{Time: time.Now()}
			updated = assets.Timestamp{Time: time.Now()}
		)

		m := mockLinkManager{
			t:        t,
			updateID: linkID,
			updateReq: assets.LinkUpdateRequest{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
			},
			updateLink: &assets.Link{
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

		updateReq := network.LinkUpdateRequest{
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var linkResp network.LinkResponse
		assert.Nil(t, json.Unmarshal(respBody, &linkResp))

		assert.Compare(t, linkResp, network.LinkResponse{Link: network.Link{
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
	linkID := assets.LinkID(uuid.New())

	t.Run("linkID failure", func(t *testing.T) {
		m := mockLinkManager{}

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid link id, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("link manager remove eailure", func(t *testing.T) {
		m := mockLinkManager{
			t:            t,
			removeLinkID: linkID,
			removeErr:    fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockLinkManager{
			t:            t,
			removeLinkID: linkID,
		}

		route := fmt.Sprintf("%s/%s", server.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// helper

func invokeLinksEndpoint(t *testing.T, m mockLinkManager, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := server.LinksService{Manager: m}
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

// mockLinkManager

type (
	mockLinkManager struct {
		t *testing.T

		filter  assets.LinksFilter
		list    []*assets.Link
		listErr error

		getID   assets.LinkID
		getLink *assets.Link
		getErr  error

		createReq  assets.LinkCreateRequest
		createLink *assets.Link
		createErr  error

		updateID   assets.LinkID
		updateReq  assets.LinkUpdateRequest
		updateLink *assets.Link
		updateErr  error

		removeLinkID assets.LinkID
		removeErr    error
	}
)

func (m mockLinkManager) List(ctx context.Context, filter assets.LinksFilter) ([]*assets.Link, error) {
	assert.Compare(m.t, filter, m.filter)
	return m.list, m.listErr
}

func (m mockLinkManager) Get(ctx context.Context, id assets.LinkID) (*assets.Link, error) {
	assert.Compare(m.t, id, m.getID)
	return m.getLink, m.getErr
}

func (m mockLinkManager) Create(ctx context.Context, req assets.LinkCreateRequest) (*assets.Link, error) {
	assert.Compare(m.t, req, m.createReq)
	return m.createLink, m.createErr
}

func (m mockLinkManager) Update(ctx context.Context, id assets.LinkID, req assets.LinkUpdateRequest) (*assets.Link, error) {
	assert.Compare(m.t, id, m.updateID)
	assert.Compare(m.t, req, m.updateReq)
	return m.updateLink, m.updateErr
}

func (m mockLinkManager) Remove(ctx context.Context, id assets.LinkID) error {
	assert.Compare(m.t, id, m.removeLinkID)
	return m.removeErr
}
