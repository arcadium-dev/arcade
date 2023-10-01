package networking_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"arcadium.dev/arcade/networking"
	"arcadium.dev/core/assert"
	"arcadium.dev/core/errors"

	"arcadium.dev/arcade"
)

func TestItemsList(t *testing.T) {
	route := networking.V1ItemsRoute
	id := uuid.New()

	t.Run("new filter failure", func(t *testing.T) {
		m := mockItemManager{}

		// ownerID failure
		w := invokeEndpoint(t, m, http.MethodGet, route, nil, "ownerID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid ownerID query parameter: 'bad uuid'")

		// locationID.ID failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid locationID query parameter: 'bad uuid'")

		// locationID.Type failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "bad type")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid locationType query parameter: 'bad type'")

		// locationID.Type missing failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String())
		assertRespError(t, w, http.StatusBadRequest, "bad request: locationType required when locationID is set")

		// owner and location failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "ownerID", id.String(), "locationID", id.String(), "locationType", "room")
		assertRespError(t, w, http.StatusBadRequest, "either ownerID or locationID/locationType can be set, not both")

		// offset failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "player", "offset", "bad offset")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid offset query parameter: 'bad offset'")

		// limit failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "item", "limit", "bad limit")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid limit query parameter: 'bad limit'")
	})

	t.Run("item manager list failure", func(t *testing.T) {
		m := mockItemManager{
			t: t,
			filter: arcade.ItemsFilter{
				LocationID: arcade.ItemID(id),
				Offset:     10,
				Limit:      10,
			},
			listErr: fmt.Errorf("%w: list failure", errors.ErrNotFound),
		}

		w := invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "item", "offset", "10", "limit", "10")
		assertRespError(t, w, http.StatusNotFound, "list failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			itemID     = arcade.ItemID(uuid.New())
			ownerID    = arcade.PlayerID(uuid.New())
			locationID = arcade.RoomID(uuid.New())
			created    = arcade.Timestamp{Time: time.Now()}
			updated    = arcade.Timestamp{Time: time.Now()}
		)

		m := mockItemManager{
			t: t,
			filter: arcade.ItemsFilter{
				LocationID: locationID,
				Offset:     25,
				Limit:      100,
			},
			list: []*arcade.Item{
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

		w := invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", locationID.ID().String(), "locationType", "room", "offset", "25", "limit", "100")

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressItems networking.EgressItems
		assert.Nil(t, json.Unmarshal(body, &egressItems))

		assert.Compare(t, egressItems, networking.EgressItems{Items: []networking.Item{
			{
				ID:          itemID.String(),
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				LocationID: networking.ItemLocationID{
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
	itemID := arcade.ItemID(uuid.New())

	t.Run("itemID failure", func(t *testing.T) {
		m := mockItemManager{}

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, "bad_itemID")

		w := invokeEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid itemID, not a well formed uuid: 'bad_itemID'")
	})

	t.Run("item manager get failure", func(t *testing.T) {
		m := mockItemManager{
			t:         t,
			getItemID: itemID,
			getErr:    fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

		w := invokeEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			ownerID    = arcade.PlayerID(uuid.New())
			locationID = arcade.PlayerID(ownerID)
			created    = arcade.Timestamp{Time: time.Now()}
			updated    = arcade.Timestamp{Time: time.Now()}
		)

		m := mockItemManager{
			t:         t,
			getItemID: itemID,
			getItem: &arcade.Item{
				ID:          itemID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			},
		}

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

		w := invokeEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressItem networking.EgressItem
		assert.Nil(t, json.Unmarshal(body, &egressItem))

		assert.Compare(t, egressItem, networking.EgressItem{Item: networking.Item{
			ID:          itemID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: networking.ItemLocationID{
				ID:   locationID.String(),
				Type: locationID.Type().String(),
			},
			Created: created,
			Updated: updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestItemCreate(t *testing.T) {
	route := networking.V1ItemsRoute

	t.Run("empty body", func(t *testing.T) {
		m := mockItemManager{}

		w := invokeEndpoint(t, m, http.MethodPost, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeEndpoint(t, m, http.MethodPost, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockItemManager{}

		w := invokeEndpoint(t, m, http.MethodPost, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("ingress item req failure", func(t *testing.T) {
		m := mockItemManager{}

		tests := []struct {
			ingressItem networking.IngressItem
			status      int
			errMsg      string
		}{
			{
				ingressItem: networking.IngressItem{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty item name",
			},
			{
				ingressItem: networking.IngressItem{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: item name exceeds maximum length",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty item description",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: item description exceeds maximum length",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID: networking.ItemLocationID{
						ID: "bad location id",
					},
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID.ID: 'bad location id', invalid UUID length: 15",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID: networking.ItemLocationID{
						ID:   uuid.New().String(),
						Type: "bad location type",
					},
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID.Type: 'bad location type'",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(test.ingressItem)
			assert.Nil(t, err)

			w := invokeEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("item manager create failure", func(t *testing.T) {
		var (
			ownerID = arcade.PlayerID(uuid.New())
			locID   = arcade.RoomID(uuid.New())
		)

		m := mockItemManager{
			t: t,
			createItemReq: arcade.IngressItem{
				Name:        "name",
				Description: "description",
				OwnerID:     arcade.PlayerID(ownerID),
				LocationID:  arcade.RoomID(locID),
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		ingressItem := networking.IngressItem{
			Name:        "name",
			Description: "description",
			OwnerID:     ownerID.String(),
			LocationID: networking.ItemLocationID{
				ID:   locID.String(),
				Type: locID.Type().String(),
			},
		}
		body, err := json.Marshal(ingressItem)
		assert.Nil(t, err)

		w := invokeEndpoint(t, m, http.MethodPost, route, body)
		assertRespError(t, w, http.StatusConflict, "conflict: create failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			itemID  = arcade.ItemID(uuid.New())
			ownerID = arcade.PlayerID(uuid.New())
			created = arcade.Timestamp{Time: time.Now()}
			updated = arcade.Timestamp{Time: time.Now()}
		)

		m := mockItemManager{
			t: t,
			createItemReq: arcade.IngressItem{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  ownerID,
			},
			createItem: &arcade.Item{
				ID:          itemID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  ownerID,
				Created:     created,
				Updated:     updated,
			},
		}

		ingressItem := networking.IngressItem{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: networking.ItemLocationID{
				ID:   ownerID.String(),
				Type: ownerID.Type().String(),
			},
		}
		body, err := json.Marshal(ingressItem)
		assert.Nil(t, err)

		w := invokeEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var itemResp networking.EgressItem
		assert.Nil(t, json.Unmarshal(respBody, &itemResp))

		assert.Compare(t, itemResp, networking.EgressItem{Item: networking.Item{
			ID:          itemID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: networking.ItemLocationID{
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
		m := mockItemManager{}

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, "bad_itemID")

		w := invokeEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid itemID, not a well formed uuid: 'bad_itemID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockItemManager{}

		itemID := uuid.New()
		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

		w := invokeEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockItemManager{}

		itemID := uuid.New()
		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

		w := invokeEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("ingress item req failure", func(t *testing.T) {
		m := mockItemManager{}

		tests := []struct {
			ingressItem networking.IngressItem
			status      int
			errMsg      string
		}{
			{
				ingressItem: networking.IngressItem{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty item name",
			},
			{
				ingressItem: networking.IngressItem{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: item name exceeds maximum length",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty item description",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: item description exceeds maximum length",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID: networking.ItemLocationID{
						ID: "bad location id",
					},
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID.ID: 'bad location id', invalid UUID length: 15",
			},
			{
				ingressItem: networking.IngressItem{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					LocationID: networking.ItemLocationID{
						ID:   uuid.New().String(),
						Type: "bad location type",
					},
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID.Type: 'bad location type'",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(test.ingressItem)
			assert.Nil(t, err)

			itemID := uuid.New()
			route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

			w := invokeEndpoint(t, m, http.MethodPut, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("item manager update failure", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			itemID  = arcade.ItemID(uuid.New())
			ownerID = arcade.PlayerID(uuid.New())
			locID   = arcade.ItemID(uuid.New())
		)

		m := mockItemManager{
			t:            t,
			updateItemID: itemID,
			updateItemReq: arcade.IngressItem{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locID,
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		ingressItem := networking.IngressItem{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: networking.ItemLocationID{
				ID:   locID.String(),
				Type: locID.Type().String(),
			},
		}
		body, err := json.Marshal(ingressItem)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

		w := invokeEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			itemID  = arcade.ItemID(uuid.New())
			ownerID = arcade.PlayerID(uuid.New())
			created = arcade.Timestamp{Time: time.Now()}
			updated = arcade.Timestamp{Time: time.Now()}
		)

		m := mockItemManager{
			t:            t,
			updateItemID: itemID,
			updateItemReq: arcade.IngressItem{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  ownerID,
			},
			updateItem: &arcade.Item{
				ID:          itemID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  ownerID,
				Created:     created,
				Updated:     updated,
			},
		}

		ingressItem := networking.IngressItem{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: networking.ItemLocationID{
				ID:   ownerID.String(),
				Type: ownerID.Type().String(),
			},
		}
		body, err := json.Marshal(ingressItem)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

		w := invokeEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var itemResp networking.EgressItem
		assert.Nil(t, json.Unmarshal(respBody, &itemResp))

		assert.Compare(t, itemResp, networking.EgressItem{Item: networking.Item{
			ID:          itemID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			LocationID: networking.ItemLocationID{
				ID:   ownerID.String(),
				Type: ownerID.Type().String(),
			},
			Created: created,
			Updated: updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestItemRemove(t *testing.T) {
	itemID := arcade.ItemID(uuid.New())

	t.Run("itemID failure", func(t *testing.T) {
		m := mockItemManager{}

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, "bad_itemID")

		w := invokeEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid itemID, not a well formed uuid: 'bad_itemID'")
	})

	t.Run("item manager remove eailure", func(t *testing.T) {
		m := mockItemManager{
			t:            t,
			removeItemID: itemID,
			removeErr:    fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

		w := invokeEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockItemManager{
			t:            t,
			removeItemID: itemID,
		}

		route := fmt.Sprintf("%s/%s", networking.V1ItemsRoute, itemID.String())

		w := invokeEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// mockItemManager

type (
	mockItemManager struct {
		t *testing.T

		filter  arcade.ItemsFilter
		list    []*arcade.Item
		listErr error

		getItemID arcade.ItemID
		getItem   *arcade.Item
		getErr    error

		createItemReq arcade.IngressItem
		createItem    *arcade.Item
		createErr     error

		updateItemID  arcade.ItemID
		updateItemReq arcade.IngressItem
		updateItem    *arcade.Item
		updateErr     error

		removeItemID arcade.ItemID
		removeErr    error
	}
)

func (m mockItemManager) List(ctx context.Context, filter arcade.ItemsFilter) ([]*arcade.Item, error) {
	assert.Compare(m.t, filter, m.filter, cmpopts.IgnoreInterfaces(struct{ arcade.ItemLocationID }{}))

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

func (m mockItemManager) Get(ctx context.Context, itemID arcade.ItemID) (*arcade.Item, error) {
	assert.Compare(m.t, itemID, m.getItemID)
	return m.getItem, m.getErr
}

func (m mockItemManager) Create(ctx context.Context, ingressItem arcade.IngressItem) (*arcade.Item, error) {
	assert.Compare(m.t, ingressItem, m.createItemReq, cmpopts.IgnoreInterfaces(struct{ arcade.ItemLocationID }{}))
	cmpIngressItem(m.t, ingressItem, m.createItemReq)
	return m.createItem, m.createErr
}

func (m mockItemManager) Update(ctx context.Context, itemID arcade.ItemID, ingressItem arcade.IngressItem) (*arcade.Item, error) {
	assert.Compare(m.t, itemID, m.updateItemID)
	assert.Compare(m.t, ingressItem, m.updateItemReq, cmpopts.IgnoreInterfaces(struct{ arcade.ItemLocationID }{}))
	cmpIngressItem(m.t, ingressItem, m.updateItemReq)
	return m.updateItem, m.updateErr
}

func (m mockItemManager) Remove(ctx context.Context, itemID arcade.ItemID) error {
	assert.Compare(m.t, itemID, m.removeItemID)
	return m.removeErr
}

func cmpIngressItem(t *testing.T, actual, expected arcade.IngressItem) {
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
