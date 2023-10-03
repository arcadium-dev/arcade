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

	"arcadium.dev/core/assert"
	"arcadium.dev/core/errors"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/networking"
)

func TestRoomsList(t *testing.T) {
	route := networking.V1RoomsRoute
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
			filter: arcade.RoomsFilter{
				ParentID: arcade.RoomID(id),
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
			roomID   = arcade.RoomID(uuid.New())
			ownerID  = arcade.PlayerID(uuid.New())
			parentID = arcade.RoomID(uuid.New())
			created  = arcade.Timestamp{Time: time.Now()}
			updated  = arcade.Timestamp{Time: time.Now()}
		)

		m := mockRoomManager{
			t: t,
			filter: arcade.RoomsFilter{
				Offset: 25,
				Limit:  100,
			},
			list: []*arcade.Room{
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

		var egressRooms networking.EgressRooms
		assert.Nil(t, json.Unmarshal(body, &egressRooms))

		assert.Compare(t, egressRooms, networking.EgressRooms{Rooms: []networking.Room{
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
	roomID := arcade.RoomID(uuid.New())

	t.Run("roomID failure", func(t *testing.T) {
		m := mockRoomManager{}

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, "bad_roomID")

		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid roomID, not a well formed uuid: 'bad_roomID'")
	})

	t.Run("room manager get failure", func(t *testing.T) {
		m := mockRoomManager{
			t:         t,
			getRoomID: roomID,
			getErr:    fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			ownerID  = arcade.PlayerID(uuid.New())
			parentID = arcade.RoomID(ownerID)
			created  = arcade.Timestamp{Time: time.Now()}
			updated  = arcade.Timestamp{Time: time.Now()}
		)

		m := mockRoomManager{
			t:         t,
			getRoomID: roomID,
			getRoom: &arcade.Room{
				ID:          roomID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			},
		}

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressRoom networking.EgressRoom
		assert.Nil(t, json.Unmarshal(body, &egressRoom))

		assert.Compare(t, egressRoom, networking.EgressRoom{Room: networking.Room{
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
	route := networking.V1RoomsRoute

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

	t.Run("ingress room req failure", func(t *testing.T) {
		m := mockRoomManager{}

		tests := []struct {
			ingressRoom networking.IngressRoom
			status      int
			errMsg      string
		}{
			{
				ingressRoom: networking.IngressRoom{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty room name",
			},
			{
				ingressRoom: networking.IngressRoom{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: room name exceeds maximum length",
			},
			{
				ingressRoom: networking.IngressRoom{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty room description",
			},
			{
				ingressRoom: networking.IngressRoom{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: room description exceeds maximum length",
			},
			{
				ingressRoom: networking.IngressRoom{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				ingressRoom: networking.IngressRoom{
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
			body, err := json.Marshal(test.ingressRoom)
			assert.Nil(t, err)

			w := invokeRoomsEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("room manager create failure", func(t *testing.T) {
		var (
			ownerID  = arcade.PlayerID(uuid.New())
			parentID = arcade.RoomID(uuid.New())
		)

		m := mockRoomManager{
			t: t,
			createRoomReq: arcade.IngressRoom{
				Name:        "name",
				Description: "description",
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		ingressRoom := networking.IngressRoom{
			Name:        "name",
			Description: "description",
			OwnerID:     ownerID.String(),
			ParentID:    parentID.String(),
		}
		body, err := json.Marshal(ingressRoom)
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
			roomID   = arcade.RoomID(uuid.New())
			ownerID  = arcade.PlayerID(uuid.New())
			parentID = arcade.RoomID(uuid.New())
			created  = arcade.Timestamp{Time: time.Now()}
			updated  = arcade.Timestamp{Time: time.Now()}
		)

		m := mockRoomManager{
			t: t,
			createRoomReq: arcade.IngressRoom{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
			createRoom: &arcade.Room{
				ID:          roomID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			},
		}

		ingressRoom := networking.IngressRoom{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			ParentID:    parentID.String(),
		}
		body, err := json.Marshal(ingressRoom)
		assert.Nil(t, err)

		w := invokeRoomsEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var roomResp networking.EgressRoom
		assert.Nil(t, json.Unmarshal(respBody, &roomResp))

		assert.Compare(t, roomResp, networking.EgressRoom{Room: networking.Room{
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

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, "bad_roomID")

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid roomID, not a well formed uuid: 'bad_roomID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockRoomManager{}

		roomID := uuid.New()
		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeRoomsEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockRoomManager{}

		roomID := uuid.New()
		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("ingress room req failure", func(t *testing.T) {
		m := mockRoomManager{}

		tests := []struct {
			ingressRoom networking.IngressRoom
			status      int
			errMsg      string
		}{
			{
				ingressRoom: networking.IngressRoom{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty room name",
			},
			{
				ingressRoom: networking.IngressRoom{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: room name exceeds maximum length",
			},
			{
				ingressRoom: networking.IngressRoom{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty room description",
			},
			{
				ingressRoom: networking.IngressRoom{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: room description exceeds maximum length",
			},
			{
				ingressRoom: networking.IngressRoom{
					Name:        randString(256),
					Description: randString(4096),
					OwnerID:     "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid ownerID: 'bad owner id'",
			},
			{
				ingressRoom: networking.IngressRoom{
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
			body, err := json.Marshal(test.ingressRoom)
			assert.Nil(t, err)

			roomID := uuid.New()
			route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

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
			roomID   = arcade.RoomID(uuid.New())
			ownerID  = arcade.PlayerID(uuid.New())
			parentID = arcade.RoomID(uuid.New())
		)

		m := mockRoomManager{
			t:            t,
			updateRoomID: roomID,
			updateRoomReq: arcade.IngressRoom{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		ingressRoom := networking.IngressRoom{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			ParentID:    parentID.String(),
		}
		body, err := json.Marshal(ingressRoom)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			roomID   = arcade.RoomID(uuid.New())
			ownerID  = arcade.PlayerID(uuid.New())
			parentID = arcade.RoomID(uuid.New())
			created  = arcade.Timestamp{Time: time.Now()}
			updated  = arcade.Timestamp{Time: time.Now()}
		)

		m := mockRoomManager{
			t:            t,
			updateRoomID: roomID,
			updateRoomReq: arcade.IngressRoom{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
			updateRoom: &arcade.Room{
				ID:          roomID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			},
		}

		ingressRoom := networking.IngressRoom{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID.String(),
			ParentID:    parentID.String(),
		}
		body, err := json.Marshal(ingressRoom)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var roomResp networking.EgressRoom
		assert.Nil(t, json.Unmarshal(respBody, &roomResp))

		assert.Compare(t, roomResp, networking.EgressRoom{Room: networking.Room{
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
	roomID := arcade.RoomID(uuid.New())

	t.Run("roomID failure", func(t *testing.T) {
		m := mockRoomManager{}

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, "bad_roomID")

		w := invokeRoomsEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid roomID, not a well formed uuid: 'bad_roomID'")
	})

	t.Run("room manager remove eailure", func(t *testing.T) {
		m := mockRoomManager{
			t:            t,
			removeRoomID: roomID,
			removeErr:    fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockRoomManager{
			t:            t,
			removeRoomID: roomID,
		}

		route := fmt.Sprintf("%s/%s", networking.V1RoomsRoute, roomID.String())

		w := invokeRoomsEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// mockRoomManager

type (
	mockRoomManager struct {
		t *testing.T

		filter  arcade.RoomsFilter
		list    []*arcade.Room
		listErr error

		getRoomID arcade.RoomID
		getRoom   *arcade.Room
		getErr    error

		createRoomReq arcade.IngressRoom
		createRoom    *arcade.Room
		createErr     error

		updateRoomID  arcade.RoomID
		updateRoomReq arcade.IngressRoom
		updateRoom    *arcade.Room
		updateErr     error

		removeRoomID arcade.RoomID
		removeErr    error
	}
)

func (m mockRoomManager) List(ctx context.Context, filter arcade.RoomsFilter) ([]*arcade.Room, error) {
	m.t.Helper()
	assert.Compare(m.t, filter, m.filter)
	return m.list, m.listErr
}

func (m mockRoomManager) Get(ctx context.Context, roomID arcade.RoomID) (*arcade.Room, error) {
	assert.Compare(m.t, roomID, m.getRoomID)
	return m.getRoom, m.getErr
}

func (m mockRoomManager) Create(ctx context.Context, ingressRoom arcade.IngressRoom) (*arcade.Room, error) {
	assert.Compare(m.t, ingressRoom, m.createRoomReq)
	return m.createRoom, m.createErr
}

func (m mockRoomManager) Update(ctx context.Context, roomID arcade.RoomID, ingressRoom arcade.IngressRoom) (*arcade.Room, error) {
	assert.Compare(m.t, roomID, m.updateRoomID)
	assert.Compare(m.t, ingressRoom, m.updateRoomReq)
	return m.updateRoom, m.updateErr
}

func (m mockRoomManager) Remove(ctx context.Context, roomID arcade.RoomID) error {
	assert.Compare(m.t, roomID, m.removeRoomID)
	return m.removeErr
}
