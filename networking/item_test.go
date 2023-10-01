package networking_test

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

	"arcadium.dev/arcade/networking"
	"arcadium.dev/core/assert"
	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade"
)

func TestItemsList(t *testing.T) {
	route := networking.V1ItemsRoute
	id := uuid.New()

	t.Run("new filter failure", func(t *testing.T) {
		m := mockItemManager{}

		// ownerID failure
		w := invokeEndpoint(t, m, http.MethodGet, route, nil, "ownerID", "bad uuid")
		checkRespError(t, w, http.StatusBadRequest, "bad request: invalid ownerID query parameter: 'bad uuid'")

		// locationID.ID failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", "bad uuid")
		checkRespError(t, w, http.StatusBadRequest, "bad request: invalid locationID query parameter: 'bad uuid'")

		// locationID.Type failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "bad type")
		checkRespError(t, w, http.StatusBadRequest, "bad request: invalid locationType query parameter: 'bad type'")

		// locationID.Type missing failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String())
		checkRespError(t, w, http.StatusBadRequest, "bad request: locationType required when locationID is set")

		// owner and location failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "ownerID", id.String(), "locationID", id.String(), "locationType", "room")
		checkRespError(t, w, http.StatusBadRequest, "either ownerID or locationID/locationType can be set, not both")

		// offset failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "player", "offset", "bad offset")
		checkRespError(t, w, http.StatusBadRequest, "bad request: invalid offset query parameter: 'bad offset'")

		// limit failure
		w = invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "item", "limit", "bad limit")
		checkRespError(t, w, http.StatusBadRequest, "bad request: invalid limit query parameter: 'bad limit'")
	})

	t.Run("item manager list failure", func(t *testing.T) {
		m := mockItemManager{
			t: t,
			filter: arcade.ItemsFilter{
				LocationID: arcade.ItemID(uuid.NullUUID{UUID: id, Valid: true}),
				Offset:     10,
				Limit:      10,
			},
			listErr: fmt.Errorf("%w: list failure", errors.ErrNotFound),
		}

		w := invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "locationType", "item", "offset", "10", "limit", "10")
		checkRespError(t, w, http.StatusNotFound, "list failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			itemID     = arcade.ItemID(uuid.NullUUID{UUID: uuid.New(), Valid: true})
			ownerID    = arcade.PlayerID(uuid.NullUUID{UUID: uuid.New(), Valid: true})
			locationID = arcade.RoomID(uuid.NullUUID{UUID: uuid.New(), Valid: true})
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

		w := invokeEndpoint(t, m, http.MethodGet, route, nil, "locationID", locationID.ID().UUID.String(), "locationType", "room", "offset", "25", "limit", "100")

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var itemsResp networking.ItemsResponse
		assert.Nil(t, json.Unmarshal(body, &itemsResp))

		assert.Compare(t, itemsResp, networking.ItemsResponse{Items: []networking.Item{
			{
				ID:          itemID.UUID.String(),
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.UUID.String(),
				LocationID: networking.ItemLocationID{
					ID:   locationID.UUID.String(),
					Type: locationID.Type().String(),
				},
				Created: created,
				Updated: updated,
			},
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

// helpers

func invokeEndpoint(t *testing.T, m mockItemManager, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := networking.ItemsService{Manager: m}
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

func checkRespError(t *testing.T, w *httptest.ResponseRecorder, status int, errMsg string) {
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, status)

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	defer resp.Body.Close()

	var respErr server.ResponseError
	err = json.Unmarshal(body, &respErr)
	assert.Nil(t, err)

	assert.Contains(t, respErr.Detail, errMsg)
	assert.Equal(t, respErr.Status, status)
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

		createItemReq arcade.ItemRequest
		createItem    *arcade.Item
		createErr     error

		updateItemID  arcade.ItemID
		updateItemReq arcade.ItemRequest
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

func (m mockItemManager) Create(ctx context.Context, itemReq arcade.ItemRequest) (*arcade.Item, error) {
	assert.Compare(m.t, itemReq, m.createItemReq)
	return m.createItem, m.createErr
}

func (m mockItemManager) Update(ctx context.Context, itemID arcade.ItemID, itemReq arcade.ItemRequest) (*arcade.Item, error) {
	assert.Compare(m.t, itemID, m.updateItemID)
	assert.Compare(m.t, itemReq, m.updateItemReq)
	return m.updateItem, m.updateErr
}

func (m mockItemManager) Remove(ctx context.Context, itemID arcade.ItemID) error {
	assert.Compare(m.t, itemID, m.removeItemID)
	return m.removeErr
}
