package networking_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"arcadium.dev/arcade/networking"
	"arcadium.dev/core/assert"
	"arcadium.dev/core/errors"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"arcadium.dev/arcade"
)

func TestLinksList(t *testing.T) {
	route := networking.V1LinksRoute
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
			filter: arcade.LinksFilter{
				LocationID: arcade.RoomID(id),
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
			linkID        = arcade.LinkID(uuid.New())
			ownerID       = arcade.PlayerID(uuid.New())
			locationID    = arcade.RoomID(uuid.New())
			destinationID = arcade.RoomID(uuid.New())
			created       = arcade.Timestamp{Time: time.Now()}
			updated       = arcade.Timestamp{Time: time.Now()}
		)

		m := mockLinkManager{
			t: t,
			filter: arcade.LinksFilter{
				Offset: 25,
				Limit:  100,
			},
			list: []*arcade.Link{
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

		var egressLinks networking.EgressLinks
		assert.Nil(t, json.Unmarshal(body, &egressLinks))

		assert.Compare(t, egressLinks, networking.EgressLinks{Links: []networking.Link{
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
	linkID := arcade.LinkID(uuid.New())

	t.Run("linkID failure", func(t *testing.T) {
		m := mockLinkManager{}

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid linkID, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("link manager get failure", func(t *testing.T) {
		m := mockLinkManager{
			t:         t,
			getLinkID: linkID,
			getErr:    fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			ownerID       = arcade.PlayerID(uuid.New())
			locationID    = arcade.RoomID(ownerID)
			destinationID = arcade.RoomID(ownerID)
			created       = arcade.Timestamp{Time: time.Now()}
			updated       = arcade.Timestamp{Time: time.Now()}
		)

		m := mockLinkManager{
			t:         t,
			getLinkID: linkID,
			getLink: &arcade.Link{
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

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressLink networking.EgressLink
		assert.Nil(t, json.Unmarshal(body, &egressLink))

		assert.Compare(t, egressLink, networking.EgressLink{Link: networking.Link{
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
	route := networking.V1LinksRoute

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

	t.Run("ingress link req failure", func(t *testing.T) {
		m := mockLinkManager{}

		tests := []struct {
			ingressLink networking.IngressLink
			status      int
			errMsg      string
		}{
			{
				ingressLink: networking.IngressLink{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link name",
			},
			{
				ingressLink: networking.IngressLink{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link name exceeds maximum length",
			},
			{
				ingressLink: networking.IngressLink{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link description",
			},
			{
				ingressLink: networking.IngressLink{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link description exceeds maximum length",
			},
			{
				ingressLink: networking.IngressLink{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				ingressLink: networking.IngressLink{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID:  "bad location id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID: 'bad location id', invalid UUID length: 15",
			},
			{
				ingressLink: networking.IngressLink{
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
			body, err := json.Marshal(test.ingressLink)
			assert.Nil(t, err)

			w := invokeLinksEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("link manager create failure", func(t *testing.T) {
		var (
			ownerID = arcade.PlayerID(uuid.New())
			locID   = arcade.RoomID(uuid.New())
			destID  = arcade.RoomID(uuid.New())
		)

		m := mockLinkManager{
			t: t,
			createLinkReq: arcade.IngressLink{
				Name:          "name",
				Description:   "description",
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		ingressLink := networking.IngressLink{
			Name:          "name",
			Description:   "description",
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
		}
		body, err := json.Marshal(ingressLink)
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
			linkID  = arcade.LinkID(uuid.New())
			ownerID = arcade.PlayerID(uuid.New())
			locID   = arcade.RoomID(uuid.New())
			destID  = arcade.RoomID(uuid.New())
			created = arcade.Timestamp{Time: time.Now()}
			updated = arcade.Timestamp{Time: time.Now()}
		)

		m := mockLinkManager{
			t: t,
			createLinkReq: arcade.IngressLink{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
			},
			createLink: &arcade.Link{
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

		ingressLink := networking.IngressLink{
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
		}
		body, err := json.Marshal(ingressLink)
		assert.Nil(t, err)

		w := invokeLinksEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var linkResp networking.EgressLink
		assert.Nil(t, json.Unmarshal(respBody, &linkResp))

		assert.Compare(t, linkResp, networking.EgressLink{Link: networking.Link{
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

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid linkID, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockLinkManager{}

		linkID := uuid.New()
		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeLinksEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockLinkManager{}

		linkID := uuid.New()
		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("ingress link req failure", func(t *testing.T) {
		m := mockLinkManager{}

		tests := []struct {
			ingressLink networking.IngressLink
			status      int
			errMsg      string
		}{
			{
				ingressLink: networking.IngressLink{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link name",
			},
			{
				ingressLink: networking.IngressLink{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link name exceeds maximum length",
			},
			{
				ingressLink: networking.IngressLink{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty link description",
			},
			{
				ingressLink: networking.IngressLink{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: link description exceeds maximum length",
			},
			{
				ingressLink: networking.IngressLink{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				ingressLink: networking.IngressLink{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID:  "bad location id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID: 'bad location id', invalid UUID length: 15",
			},
			{
				ingressLink: networking.IngressLink{
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
			body, err := json.Marshal(test.ingressLink)
			assert.Nil(t, err)

			linkID := uuid.New()
			route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

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
			linkID  = arcade.LinkID(uuid.New())
			ownerID = arcade.PlayerID(uuid.New())
			locID   = arcade.RoomID(uuid.New())
			destID  = arcade.RoomID(uuid.New())
		)

		m := mockLinkManager{
			t:            t,
			updateLinkID: linkID,
			updateLinkReq: arcade.IngressLink{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		ingressLink := networking.IngressLink{
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
		}
		body, err := json.Marshal(ingressLink)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			linkID  = arcade.LinkID(uuid.New())
			ownerID = arcade.PlayerID(uuid.New())
			locID   = arcade.RoomID(uuid.New())
			destID  = arcade.RoomID(uuid.New())
			created = arcade.Timestamp{Time: time.Now()}
			updated = arcade.Timestamp{Time: time.Now()}
		)

		m := mockLinkManager{
			t:            t,
			updateLinkID: linkID,
			updateLinkReq: arcade.IngressLink{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locID,
				DestinationID: destID,
			},
			updateLink: &arcade.Link{
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

		ingressLink := networking.IngressLink{
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID.String(),
			LocationID:    locID.String(),
			DestinationID: destID.String(),
		}
		body, err := json.Marshal(ingressLink)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var linkResp networking.EgressLink
		assert.Nil(t, json.Unmarshal(respBody, &linkResp))

		assert.Compare(t, linkResp, networking.EgressLink{Link: networking.Link{
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
	linkID := arcade.LinkID(uuid.New())

	t.Run("linkID failure", func(t *testing.T) {
		m := mockLinkManager{}

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, "bad_linkID")

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid linkID, not a well formed uuid: 'bad_linkID'")
	})

	t.Run("link manager remove eailure", func(t *testing.T) {
		m := mockLinkManager{
			t:            t,
			removeLinkID: linkID,
			removeErr:    fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockLinkManager{
			t:            t,
			removeLinkID: linkID,
		}

		route := fmt.Sprintf("%s/%s", networking.V1LinksRoute, linkID.String())

		w := invokeLinksEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// mockLinkManager

type (
	mockLinkManager struct {
		t *testing.T

		filter  arcade.LinksFilter
		list    []*arcade.Link
		listErr error

		getLinkID arcade.LinkID
		getLink   *arcade.Link
		getErr    error

		createLinkReq arcade.IngressLink
		createLink    *arcade.Link
		createErr     error

		updateLinkID  arcade.LinkID
		updateLinkReq arcade.IngressLink
		updateLink    *arcade.Link
		updateErr     error

		removeLinkID arcade.LinkID
		removeErr    error
	}
)

func (m mockLinkManager) List(ctx context.Context, filter arcade.LinksFilter) ([]*arcade.Link, error) {
	assert.Compare(m.t, filter, m.filter)
	return m.list, m.listErr
}

func (m mockLinkManager) Get(ctx context.Context, linkID arcade.LinkID) (*arcade.Link, error) {
	assert.Compare(m.t, linkID, m.getLinkID)
	return m.getLink, m.getErr
}

func (m mockLinkManager) Create(ctx context.Context, ingressLink arcade.IngressLink) (*arcade.Link, error) {
	assert.Compare(m.t, ingressLink, m.createLinkReq)
	return m.createLink, m.createErr
}

func (m mockLinkManager) Update(ctx context.Context, linkID arcade.LinkID, ingressLink arcade.IngressLink) (*arcade.Link, error) {
	assert.Compare(m.t, linkID, m.updateLinkID)
	assert.Compare(m.t, ingressLink, m.updateLinkReq)
	return m.updateLink, m.updateErr
}

func (m mockLinkManager) Remove(ctx context.Context, linkID arcade.LinkID) error {
	assert.Compare(m.t, linkID, m.removeLinkID)
	return m.removeErr
}
