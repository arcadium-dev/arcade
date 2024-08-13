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

func TestItemsList(t *testing.T) {
	route := server.V1ItemRoute
	id := uuid.New()

	t.Run("new filter failure", func(t *testing.T) {
		m := mockItemStorage{}

		// ownerID failure
		w := invokeItemsEndpoint(t, m, http.MethodGet, route, nil, "ownerID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid ownerID query parameter: 'bad uuid'")

		// locationID.ID failure
		w = invokeItemsEndpoint(t, m, http.MethodGet, route, nil, "locationID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid locationID query parameter: 'bad uuid'")

		// locationID.Type failure
		w = invokeItemsEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "bad type")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid locationType query parameter: 'bad type'")

		// locationID.Type missing failure
		w = invokeItemsEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String())
		assertRespError(t, w, http.StatusBadRequest, "bad request: locationType required when locationID is set")

		// offset failure
		w = invokeItemsEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "player", "offset", "bad offset")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid offset query parameter: 'bad offset'")

		// limit failure
		w = invokeItemsEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "item", "limit", "bad limit")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid limit query parameter: 'bad limit'")
	})

	t.Run("item manager list failure", func(t *testing.T) {
		m := mockItemStorage{
			t: t,
			filter: asset.ItemFilter{
				LocationID: asset.ItemID(id),
				Offset:     10,
				Limit:      10,
			},
			listErr: fmt.Errorf("%w: list failure", errors.ErrNotFound),
		}

		w := invokeItemsEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "item", "offset", "10", "limit", "10")
		assertRespError(t, w, http.StatusNotFound, "list failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			itemID     = asset.ItemID(uuid.New())
			ownerID    = asset.PlayerID(uuid.New())
			locationID = asset.RoomID(uuid.New())
			created    = arcade.Timestamp{Time: time.Now()}
			updated    = arcade.Timestamp{Time: time.Now()}
		)

		m := mockItemStorage{
			t: t,
			filter: asset.ItemFilter{
				LocationID: locationID,
				Offset:     25,
				Limit:      100,
			},
			list: []*asset.Item{
				{
					ID:          itemID,
					Name:        name,
					Description: desc,
					OwnerID:     ownerID,
					LocationID:  locationID,
					Created:     created,
					Updated:     updated,
				},
			},
		}

		w := invokeItemsEndpoint(t, m, http.MethodGet, route, nil, "locationID", locationID.ID().String(), "locationType", "room", "offset", "25", "limit", "100")

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var itemsResp rest.ItemsResponse
		assert.Nil(t, json.Unmarshal(body, &itemsResp))

		assert.Compare(t, itemsResp, rest.ItemsResponse{Items: []rest.Item{
			{
				ID:          itemID.String(),
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				LocationID: rest.ItemLocationID{
					ID:   locationID.String(),
					Type: locationID.Type().String(),
				},
				Created: created,
				Updated: updated,
			},
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestItemGet(t *testing.T) {
	itemID := asset.ItemID(uuid.New())

	t.Run("itemID failure", func(t *testing.T) {
		m := mockItemStorage{}

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, "bad_itemID")

		w := invokeItemsEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid item id, not a well formed uuid: 'bad_itemID'")
	})

	t.Run("item manager get failure", func(t *testing.T) {
		m := mockItemStorage{
			t:      t,
			getID:  itemID,
			getErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

		w := invokeItemsEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			ownerID    = asset.PlayerID(uuid.New())
			locationID = asset.PlayerID(ownerID)
			created    = arcade.Timestamp{Time: time.Now()}
			updated    = arcade.Timestamp{Time: time.Now()}
		)

		m := mockItemStorage{
			t:     t,
			getID: itemID,
			getItem: &asset.Item{
				ID:          itemID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			},
		}

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

		w := invokeItemsEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var itemResp rest.ItemResponse
		assert.Nil(t, json.Unmarshal(body, &itemResp))

		assert.Compare(t, itemResp, rest.ItemResponse{Item: rest.Item{
			ID:          itemID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: rest.ItemLocationID{
				ID:   locationID.String(),
				Type: locationID.Type().String(),
			},
			Created: created,
			Updated: updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestItemCreate(t *testing.T) {
	route := server.V1ItemRoute

	t.Run("empty body", func(t *testing.T) {
		m := mockItemStorage{}

		w := invokeItemsEndpoint(t, m, http.MethodPost, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeItemsEndpoint(t, m, http.MethodPost, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockItemStorage{}

		w := invokeItemsEndpoint(t, m, http.MethodPost, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("create item req failure", func(t *testing.T) {
		m := mockItemStorage{}

		tests := []struct {
			req    rest.ItemRequest
			status int
			errMsg string
		}{
			{
				req: rest.ItemRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty item name",
			},
			{
				req: rest.ItemRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: item name exceeds maximum length",
			},
			{
				req: rest.ItemRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty item description",
			},
			{
				req: rest.ItemRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: item description exceeds maximum length",
			},
			{
				req: rest.ItemRequest{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				req: rest.ItemRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID: rest.ItemLocationID{
						ID: "bad location id",
					},
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID.ID: 'bad location id', invalid UUID length: 15",
			},
			{
				req: rest.ItemRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID: rest.ItemLocationID{
						ID:   uuid.New().String(),
						Type: "bad location type",
					},
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID.Type: 'bad location type'",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.ItemCreateRequest{ItemRequest: test.req})
			assert.Nil(t, err)

			w := invokeItemsEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("item manager create failure", func(t *testing.T) {
		var (
			ownerID = asset.PlayerID(uuid.New())
			locID   = asset.RoomID(uuid.New())
		)

		m := mockItemStorage{
			t: t,
			create: asset.ItemCreate{
				ItemChange: asset.ItemChange{
					Name:        "name",
					Description: "description",
					OwnerID:     ownerID,
					LocationID:  locID,
				},
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		create := rest.ItemCreateRequest{
			ItemRequest: rest.ItemRequest{
				Name:        "name",
				Description: "description",
				OwnerID:     ownerID.String(),
				LocationID: rest.ItemLocationID{
					ID:   locID.String(),
					Type: locID.Type().String(),
				},
			},
		}
		body, err := json.Marshal(create)
		assert.Nil(t, err)

		w := invokeItemsEndpoint(t, m, http.MethodPost, route, body)
		assertRespError(t, w, http.StatusConflict, "conflict: create failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			itemID  = asset.ItemID(uuid.New())
			ownerID = asset.PlayerID(uuid.New())
			created = arcade.Timestamp{Time: time.Now()}
			updated = arcade.Timestamp{Time: time.Now()}
		)

		m := mockItemStorage{
			t: t,
			create: asset.ItemCreate{
				ItemChange: asset.ItemChange{
					Name:        name,
					Description: desc,
					OwnerID:     ownerID,
					LocationID:  ownerID,
				},
			},
			createItem: &asset.Item{
				ID:          itemID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  ownerID,
				Created:     created,
				Updated:     updated,
			},
		}

		createReq := rest.ItemCreateRequest{
			ItemRequest: rest.ItemRequest{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				LocationID: rest.ItemLocationID{
					ID:   ownerID.String(),
					Type: ownerID.Type().String(),
				},
			},
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokeItemsEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var itemResp rest.ItemResponse
		assert.Nil(t, json.Unmarshal(respBody, &itemResp))

		assert.Compare(t, itemResp, rest.ItemResponse{Item: rest.Item{
			ID:          itemID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: rest.ItemLocationID{
				ID:   ownerID.String(),
				Type: "player",
			},
			Created: created,
			Updated: updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestItemUpdate(t *testing.T) {
	t.Run("itemID failure", func(t *testing.T) {
		m := mockItemStorage{}

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, "bad_itemID")

		w := invokeItemsEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid item id, not a well formed uuid: 'bad_itemID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockItemStorage{}

		itemID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

		w := invokeItemsEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeItemsEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockItemStorage{}

		itemID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

		w := invokeItemsEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("update item req failure", func(t *testing.T) {
		m := mockItemStorage{}

		tests := []struct {
			req    rest.ItemRequest
			status int
			errMsg string
		}{
			{
				req: rest.ItemRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty item name",
			},
			{
				req: rest.ItemRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: item name exceeds maximum length",
			},
			{
				req: rest.ItemRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty item description",
			},
			{
				req: rest.ItemRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: item description exceeds maximum length",
			},
			{
				req: rest.ItemRequest{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				req: rest.ItemRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID: rest.ItemLocationID{
						ID: "bad location id",
					},
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID.ID: 'bad location id', invalid UUID length: 15",
			},
			{
				req: rest.ItemRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID: rest.ItemLocationID{
						ID:   uuid.New().String(),
						Type: "bad location type",
					},
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID.Type: 'bad location type'",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.ItemUpdateRequest{ItemRequest: test.req})
			assert.Nil(t, err)

			itemID := uuid.New()
			route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

			w := invokeItemsEndpoint(t, m, http.MethodPut, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("item manager update failure", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			itemID  = asset.ItemID(uuid.New())
			ownerID = asset.PlayerID(uuid.New())
			locID   = asset.ItemID(uuid.New())
		)

		m := mockItemStorage{
			t:        t,
			updateID: itemID,
			update: asset.ItemUpdate{
				ItemChange: asset.ItemChange{
					Name:        name,
					Description: desc,
					OwnerID:     ownerID,
					LocationID:  locID,
				},
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		updateReq := rest.ItemUpdateRequest{
			ItemRequest: rest.ItemRequest{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				LocationID: rest.ItemLocationID{
					ID:   locID.String(),
					Type: locID.Type().String(),
				},
			},
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

		w := invokeItemsEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			itemID  = asset.ItemID(uuid.New())
			ownerID = asset.PlayerID(uuid.New())
			created = arcade.Timestamp{Time: time.Now()}
			updated = arcade.Timestamp{Time: time.Now()}
		)

		m := mockItemStorage{
			t:        t,
			updateID: itemID,
			update: asset.ItemUpdate{
				ItemChange: asset.ItemChange{
					Name:        name,
					Description: desc,
					OwnerID:     ownerID,
					LocationID:  ownerID,
				},
			},
			updateItem: &asset.Item{
				ID:          itemID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  ownerID,
				Created:     created,
				Updated:     updated,
			},
		}

		updateReq := rest.ItemUpdateRequest{
			ItemRequest: rest.ItemRequest{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				LocationID: rest.ItemLocationID{
					ID:   ownerID.String(),
					Type: ownerID.Type().String(),
				},
			},
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

		w := invokeItemsEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var itemResp rest.ItemResponse
		assert.Nil(t, json.Unmarshal(respBody, &itemResp))

		assert.Compare(t, itemResp, rest.ItemResponse{Item: rest.Item{
			ID:          itemID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: rest.ItemLocationID{
				ID:   ownerID.String(),
				Type: ownerID.Type().String(),
			},
			Created: created,
			Updated: updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestItemRemove(t *testing.T) {
	itemID := asset.ItemID(uuid.New())

	t.Run("itemID failure", func(t *testing.T) {
		m := mockItemStorage{}

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, "bad_itemID")

		w := invokeItemsEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid item id, not a well formed uuid: 'bad_itemID'")
	})

	t.Run("item manager remove eailure", func(t *testing.T) {
		m := mockItemStorage{
			t:         t,
			removeID:  itemID,
			removeErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

		w := invokeItemsEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockItemStorage{
			t:        t,
			removeID: itemID,
		}

		route := fmt.Sprintf("%s/%s", server.V1ItemRoute, itemID.String())

		w := invokeItemsEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// helper

func invokeItemsEndpoint(t *testing.T, m mockItemStorage, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := server.ItemsService{Storage: m}
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

// mockItemStorage

type (
	mockItemStorage struct {
		t *testing.T

		filter  asset.ItemFilter
		list    []*asset.Item
		listErr error

		getID   asset.ItemID
		getItem *asset.Item
		getErr  error

		create     asset.ItemCreate
		createItem *asset.Item
		createErr  error

		updateID   asset.ItemID
		update     asset.ItemUpdate
		updateItem *asset.Item
		updateErr  error

		removeID  asset.ItemID
		removeErr error
	}
)

func (m mockItemStorage) List(ctx context.Context, filter asset.ItemFilter) ([]*asset.Item, error) {
	assert.Compare(m.t, filter, m.filter, cmpopts.IgnoreInterfaces(struct{ asset.ItemLocationID }{}))

	if filter.LocationID == nil && m.filter.LocationID != nil {
		m.t.Errorf("Failed: locationID mismatch, present in actual, missing in expected")
	}
	if filter.LocationID != nil && m.filter.LocationID == nil {
		m.t.Errorf("Failed: locationID mismatch, missing in actual, present in expected")
	}
	if filter.LocationID != nil {
		assert.Equal(m.t, filter.LocationID.Type(), m.filter.LocationID.Type())
		assert.Compare(m.t, filter.LocationID.ID(), m.filter.LocationID.ID())
	}

	return m.list, m.listErr
}

func (m mockItemStorage) Get(ctx context.Context, id asset.ItemID) (*asset.Item, error) {
	assert.Compare(m.t, id, m.getID)
	return m.getItem, m.getErr
}

func (m mockItemStorage) Create(ctx context.Context, create asset.ItemCreate) (*asset.Item, error) {
	assert.Compare(m.t, create, m.create, cmpopts.IgnoreInterfaces(struct{ asset.ItemLocationID }{}))
	cmpItemRequest(m.t, create.ItemChange, m.create.ItemChange)
	return m.createItem, m.createErr
}

func (m mockItemStorage) Update(ctx context.Context, id asset.ItemID, update asset.ItemUpdate) (*asset.Item, error) {
	assert.Compare(m.t, id, m.updateID)
	assert.Compare(m.t, update, m.update, cmpopts.IgnoreInterfaces(struct{ asset.ItemLocationID }{}))
	cmpItemRequest(m.t, update.ItemChange, m.update.ItemChange)
	return m.updateItem, m.updateErr
}

func (m mockItemStorage) Remove(ctx context.Context, id asset.ItemID) error {
	assert.Compare(m.t, id, m.removeID)
	return m.removeErr
}

func cmpItemRequest(t *testing.T, actual, expected asset.ItemChange) {
	t.Helper()

	if actual.LocationID == nil && expected.LocationID != nil {
		t.Errorf("Failed: locationID mismatch, present in actual, missing in expected")
	}
	if actual.LocationID != nil && expected.LocationID == nil {
		t.Errorf("Failed: locationID mismatch, missing in actual, present in expected")
	}
	if actual.LocationID != nil {
		assert.Equal(t, actual.LocationID.Type(), expected.LocationID.Type())
		assert.Compare(t, actual.LocationID.ID(), expected.LocationID.ID())
	}
}
