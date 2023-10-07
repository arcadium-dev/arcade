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
	"arcadium.dev/arcade/assets/network/rest"
	"arcadium.dev/arcade/assets/network/rest/server"
)

func TestRoomsList(t *testing.T) {
	route := server.V1RoomsRoute
	id := uuid.New()

	t.Run("new filter failure", func(t *testing.T) {
		m := mockRoomManager{}

		// ownerID failure
		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil, "ownerID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid ownerID query parameter: 'bad uuid'")

		// parentID failure
		w = invokeRoomsEndpoint(t, m, http.MethodGet, route, nil, "parentID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid parentID query parameter: 'bad uuid'")

		// offset failure
		w = invokeRoomsEndpoint(t, m, http.MethodGet, route, nil, "offset", "bad offset")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid offset query parameter: 'bad offset'")

		// limit failure
		w = invokeRoomsEndpoint(t, m, http.MethodGet, route, nil, "limit", "bad limit")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid limit query parameter: 'bad limit'")
	})

	t.Run("room manager list failure", func(t *testing.T) {
		m := mockRoomManager{
			t: t,
			filter: assets.RoomsFilter{
				ParentID: assets.RoomID(id),
				Offset:   10,
				Limit:    10,
			},
			listErr: fmt.Errorf("%w: list failure", errors.ErrNotFound),
		}

		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil, "parentID", id.String(), "offset", "10", "limit", "10")
		assertRespError(t, w, http.StatusNotFound, "list failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			roomID   = assets.RoomID(uuid.New())
			ownerID  = assets.PlayerID(uuid.New())
			parentID = assets.RoomID(uuid.New())
			created  = assets.Timestamp{Time: time.Now()}
			updated  = assets.Timestamp{Time: time.Now()}
		)

		m := mockRoomManager{
			t: t,
			filter: assets.RoomsFilter{
				Offset: 25,
				Limit:  100,
			},
			list: []*assets.Room{
				{
					ID:          roomID,
					Name:        name,
					Description: desc,
					OwnerID:     ownerID,
					ParentID:    parentID,
					Created:     created,
					Updated:     updated,
				},
			},
		}

		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil, "offset", "25", "limit", "100")

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var roomsResp rest.RoomsResponse
		assert.Nil(t, json.Unmarshal(body, &roomsResp))

		assert.Compare(t, roomsResp, rest.RoomsResponse{Rooms: []rest.Room{
			{
				ID:          roomID.String(),
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				ParentID:    parentID.String(),
				Created:     created,
				Updated:     updated,
			},
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestRoomGet(t *testing.T) {
	roomID := assets.RoomID(uuid.New())

	t.Run("roomID failure", func(t *testing.T) {
		m := mockRoomManager{}

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, "bad_roomID")

		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid room id, not a well formed uuid: 'bad_roomID'")
	})

	t.Run("room manager get failure", func(t *testing.T) {
		m := mockRoomManager{
			t:      t,
			getID:  roomID,
			getErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			ownerID  = assets.PlayerID(uuid.New())
			parentID = assets.RoomID(ownerID)
			created  = assets.Timestamp{Time: time.Now()}
			updated  = assets.Timestamp{Time: time.Now()}
		)

		m := mockRoomManager{
			t:     t,
			getID: roomID,
			getRoom: &assets.Room{
				ID:          roomID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			},
		}

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var roomResp rest.RoomResponse
		assert.Nil(t, json.Unmarshal(body, &roomResp))

		assert.Compare(t, roomResp, rest.RoomResponse{Room: rest.Room{
			ID:          roomID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			ParentID:    parentID.String(),
			Created:     created,
			Updated:     updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestRoomCreate(t *testing.T) {
	route := server.V1RoomsRoute

	t.Run("empty body", func(t *testing.T) {
		m := mockRoomManager{}

		w := invokeRoomsEndpoint(t, m, http.MethodPost, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeRoomsEndpoint(t, m, http.MethodPost, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockRoomManager{}

		w := invokeRoomsEndpoint(t, m, http.MethodPost, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("create room req failure", func(t *testing.T) {
		m := mockRoomManager{}

		tests := []struct {
			req    rest.RoomRequest
			status int
			errMsg string
		}{
			{
				req: rest.RoomRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty room name",
			},
			{
				req: rest.RoomRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: room name exceeds maximum length",
			},
			{
				req: rest.RoomRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty room description",
			},
			{
				req: rest.RoomRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: room description exceeds maximum length",
			},
			{
				req: rest.RoomRequest{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				req: rest.RoomRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					ParentID:    "bad parent id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid parentID: 'bad parent id', invalid UUID length: 13",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.RoomCreateRequest{RoomRequest: test.req})
			assert.Nil(t, err)

			w := invokeRoomsEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("room manager create failure", func(t *testing.T) {
		var (
			ownerID  = assets.PlayerID(uuid.New())
			parentID = assets.RoomID(uuid.New())
		)

		m := mockRoomManager{
			t: t,
			create: assets.RoomCreate{
				RoomChange: assets.RoomChange{
					Name:        "name",
					Description: "description",
					OwnerID:     ownerID,
					ParentID:    parentID,
				},
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		createReq := rest.RoomCreateRequest{
			RoomRequest: rest.RoomRequest{
				Name:        "name",
				Description: "description",
				OwnerID:     ownerID.String(),
				ParentID:    parentID.String(),
			},
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokeRoomsEndpoint(t, m, http.MethodPost, route, body)
		assertRespError(t, w, http.StatusConflict, "conflict: create failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			roomID   = assets.RoomID(uuid.New())
			ownerID  = assets.PlayerID(uuid.New())
			parentID = assets.RoomID(uuid.New())
			created  = assets.Timestamp{Time: time.Now()}
			updated  = assets.Timestamp{Time: time.Now()}
		)

		m := mockRoomManager{
			t: t,
			create: assets.RoomCreate{
				RoomChange: assets.RoomChange{
					Name:        name,
					Description: desc,
					OwnerID:     ownerID,
					ParentID:    parentID,
				},
			},
			createRoom: &assets.Room{
				ID:          roomID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			},
		}

		createReq := rest.RoomCreateRequest{
			RoomRequest: rest.RoomRequest{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				ParentID:    parentID.String(),
			},
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokeRoomsEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var roomResp rest.RoomResponse
		assert.Nil(t, json.Unmarshal(respBody, &roomResp))

		assert.Compare(t, roomResp, rest.RoomResponse{Room: rest.Room{
			ID:          roomID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			ParentID:    parentID.String(),
			Created:     created,
			Updated:     updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestRoomUpdate(t *testing.T) {
	t.Run("roomID failure", func(t *testing.T) {
		m := mockRoomManager{}

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, "bad_roomID")

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid room id, not a well formed uuid: 'bad_roomID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockRoomManager{}

		roomID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeRoomsEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockRoomManager{}

		roomID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("update room req failure", func(t *testing.T) {
		m := mockRoomManager{}

		tests := []struct {
			req    rest.RoomRequest
			status int
			errMsg string
		}{
			{
				req: rest.RoomRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty room name",
			},
			{
				req: rest.RoomRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: room name exceeds maximum length",
			},
			{
				req: rest.RoomRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty room description",
			},
			{
				req: rest.RoomRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: room description exceeds maximum length",
			},
			{
				req: rest.RoomRequest{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				req: rest.RoomRequest{
					Name:        "name",
					Description: "description",
					OwnerID:     uuid.New().String(),
					ParentID:    "bad parent id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid parentID: 'bad parent id', invalid UUID length: 13",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.RoomUpdateRequest{RoomRequest: test.req})
			assert.Nil(t, err)

			roomID := uuid.New()
			route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

			w := invokeRoomsEndpoint(t, m, http.MethodPut, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("room manager update failure", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			roomID   = assets.RoomID(uuid.New())
			ownerID  = assets.PlayerID(uuid.New())
			parentID = assets.RoomID(uuid.New())
		)

		m := mockRoomManager{
			t:        t,
			updateID: roomID,
			update: assets.RoomUpdate{
				RoomChange: assets.RoomChange{
					Name:        name,
					Description: desc,
					OwnerID:     ownerID,
					ParentID:    parentID,
				},
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		updateReq := rest.RoomUpdateRequest{
			RoomRequest: rest.RoomRequest{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				ParentID:    parentID.String(),
			},
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			roomID   = assets.RoomID(uuid.New())
			ownerID  = assets.PlayerID(uuid.New())
			parentID = assets.RoomID(uuid.New())
			created  = assets.Timestamp{Time: time.Now()}
			updated  = assets.Timestamp{Time: time.Now()}
		)

		m := mockRoomManager{
			t:        t,
			updateID: roomID,
			update: assets.RoomUpdate{
				RoomChange: assets.RoomChange{
					Name:        name,
					Description: desc,
					OwnerID:     ownerID,
					ParentID:    parentID,
				},
			},
			updateRoom: &assets.Room{
				ID:          roomID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			},
		}

		updateReq := rest.RoomUpdateRequest{
			RoomRequest: rest.RoomRequest{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID.String(),
				ParentID:    parentID.String(),
			},
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var roomResp rest.RoomResponse
		assert.Nil(t, json.Unmarshal(respBody, &roomResp))

		assert.Compare(t, roomResp, rest.RoomResponse{Room: rest.Room{
			ID:          roomID.String(),
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			ParentID:    parentID.String(),
			Created:     created,
			Updated:     updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestRoomRemove(t *testing.T) {
	roomID := assets.RoomID(uuid.New())

	t.Run("roomID failure", func(t *testing.T) {
		m := mockRoomManager{}

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, "bad_roomID")

		w := invokeRoomsEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid room id, not a well formed uuid: 'bad_roomID'")
	})

	t.Run("room manager remove eailure", func(t *testing.T) {
		m := mockRoomManager{
			t:         t,
			removeID:  roomID,
			removeErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockRoomManager{
			t:        t,
			removeID: roomID,
		}

		route := fmt.Sprintf("%s/%s", server.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// helper

func invokeRoomsEndpoint(t *testing.T, m mockRoomManager, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := server.RoomsService{Manager: m}
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

// mockRoomManager

type (
	mockRoomManager struct {
		t *testing.T

		filter  assets.RoomsFilter
		list    []*assets.Room
		listErr error

		getID   assets.RoomID
		getRoom *assets.Room
		getErr  error

		create     assets.RoomCreate
		createRoom *assets.Room
		createErr  error

		updateID   assets.RoomID
		update     assets.RoomUpdate
		updateRoom *assets.Room
		updateErr  error

		removeID  assets.RoomID
		removeErr error
	}
)

func (m mockRoomManager) List(ctx context.Context, filter assets.RoomsFilter) ([]*assets.Room, error) {
	m.t.Helper()
	assert.Compare(m.t, filter, m.filter)
	return m.list, m.listErr
}

func (m mockRoomManager) Get(ctx context.Context, id assets.RoomID) (*assets.Room, error) {
	assert.Compare(m.t, id, m.getID)
	return m.getRoom, m.getErr
}

func (m mockRoomManager) Create(ctx context.Context, create assets.RoomCreate) (*assets.Room, error) {
	assert.Compare(m.t, create, m.create)
	return m.createRoom, m.createErr
}

func (m mockRoomManager) Update(ctx context.Context, id assets.RoomID, update assets.RoomUpdate) (*assets.Room, error) {
	assert.Compare(m.t, id, m.updateID)
	assert.Compare(m.t, update, m.update)
	return m.updateRoom, m.updateErr
}

func (m mockRoomManager) Remove(ctx context.Context, id assets.RoomID) error {
	assert.Compare(m.t, id, m.removeID)
	return m.removeErr
}
