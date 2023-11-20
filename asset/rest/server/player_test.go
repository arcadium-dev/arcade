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

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
	"arcadium.dev/arcade/asset/rest/server"
)

func TestPlayersList(t *testing.T) {
	route := server.V1PlayerRoute
	id := uuid.New()

	t.Run("new filter failure", func(t *testing.T) {
		m := mockPlayerStorage{}

		// locationID failure
		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil, "locationID", "bad uuid")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid locationID query parameter: 'bad uuid'")

		// offset failure
		w = invokePlayersEndpoint(t, m, http.MethodGet, route, nil, "offset", "bad offset")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid offset query parameter: 'bad offset'")

		// limit failure
		w = invokePlayersEndpoint(t, m, http.MethodGet, route, nil, "limit", "bad limit")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid limit query parameter: 'bad limit'")
	})

	t.Run("player manager list failure", func(t *testing.T) {
		m := mockPlayerStorage{
			t: t,
			filter: asset.PlayerFilter{
				LocationID: asset.RoomID(id),
				Offset:     10,
				Limit:      10,
			},
			listErr: fmt.Errorf("%w: list failure", errors.ErrNotFound),
		}

		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil, "locationID", id.String(), "offset", "10", "limit", "10")
		assertRespError(t, w, http.StatusNotFound, "list failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			playerID   = asset.PlayerID(uuid.New())
			homeID     = asset.RoomID(uuid.New())
			locationID = asset.RoomID(uuid.New())
			created    = asset.Timestamp{Time: time.Now()}
			updated    = asset.Timestamp{Time: time.Now()}
		)

		m := mockPlayerStorage{
			t: t,
			filter: asset.PlayerFilter{
				Offset: 25,
				Limit:  100,
			},
			list: []*asset.Player{
				{
					ID:          playerID,
					Name:        name,
					Description: desc,
					HomeID:      homeID,
					LocationID:  locationID,
					Created:     created,
					Updated:     updated,
				},
			},
		}

		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil, "offset", "25", "limit", "100")

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressPlayers rest.PlayersResponse
		assert.Nil(t, json.Unmarshal(body, &egressPlayers))

		assert.Compare(t, egressPlayers, rest.PlayersResponse{Players: []rest.Player{
			{
				ID:          playerID.String(),
				Name:        name,
				Description: desc,
				HomeID:      homeID.String(),
				LocationID:  locationID.String(),
				Created:     created,
				Updated:     updated,
			},
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestPlayerGet(t *testing.T) {
	playerID := asset.PlayerID(uuid.New())

	t.Run("playerID failure", func(t *testing.T) {
		m := mockPlayerStorage{}

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, "bad_playerID")

		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid player id, not a well formed uuid: 'bad_playerID'")
	})

	t.Run("player manager get failure", func(t *testing.T) {
		m := mockPlayerStorage{
			t:      t,
			getID:  playerID,
			getErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			homeID     = asset.RoomID(uuid.New())
			locationID = asset.RoomID(homeID)
			created    = asset.Timestamp{Time: time.Now()}
			updated    = asset.Timestamp{Time: time.Now()}
		)

		m := mockPlayerStorage{
			t:     t,
			getID: playerID,
			getPlayer: &asset.Player{
				ID:          playerID,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			},
		}

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var playerResp rest.PlayerResponse
		assert.Nil(t, json.Unmarshal(body, &playerResp))

		assert.Compare(t, playerResp, rest.PlayerResponse{Player: rest.Player{
			ID:          playerID.String(),
			Name:        name,
			Description: desc,
			HomeID:      homeID.String(),
			LocationID:  locationID.String(),
			Created:     created,
			Updated:     updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestPlayerCreate(t *testing.T) {
	route := server.V1PlayerRoute

	t.Run("empty body", func(t *testing.T) {
		m := mockPlayerStorage{}

		w := invokePlayersEndpoint(t, m, http.MethodPost, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokePlayersEndpoint(t, m, http.MethodPost, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockPlayerStorage{}

		w := invokePlayersEndpoint(t, m, http.MethodPost, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("create player req failure", func(t *testing.T) {
		m := mockPlayerStorage{}

		tests := []struct {
			req    rest.PlayerRequest
			status int
			errMsg string
		}{
			{
				req: rest.PlayerRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty player name",
			},
			{
				req: rest.PlayerRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: player name exceeds maximum length",
			},
			{
				req: rest.PlayerRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty player description",
			},
			{
				req: rest.PlayerRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: player description exceeds maximum length",
			},
			{
				req: rest.PlayerRequest{
					Name:        randString(256),
					Description: randString(4096),
					HomeID:      "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid homeID: 'bad owner id'",
			},
			{
				req: rest.PlayerRequest{
					Name:        "name",
					Description: "description",
					HomeID:      uuid.New().String(),
					LocationID:  "bad location id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID: 'bad location id', invalid UUID length: 15",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.PlayerCreateRequest{PlayerRequest: test.req})
			assert.Nil(t, err)

			w := invokePlayersEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("player manager create failure", func(t *testing.T) {
		var (
			homeID = asset.RoomID(uuid.New())
			locID  = asset.RoomID(uuid.New())
		)

		m := mockPlayerStorage{
			t: t,
			create: asset.PlayerCreate{
				PlayerChange: asset.PlayerChange{
					Name:        "name",
					Description: "description",
					HomeID:      homeID,
					LocationID:  locID,
				},
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		createReq := rest.PlayerCreateRequest{
			PlayerRequest: rest.PlayerRequest{
				Name:        "name",
				Description: "description",
				HomeID:      homeID.String(),
				LocationID:  locID.String(),
			},
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokePlayersEndpoint(t, m, http.MethodPost, route, body)
		assertRespError(t, w, http.StatusConflict, "conflict: create failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			playerID = asset.PlayerID(uuid.New())
			homeID   = asset.RoomID(uuid.New())
			locID    = asset.RoomID(uuid.New())
			created  = asset.Timestamp{Time: time.Now()}
			updated  = asset.Timestamp{Time: time.Now()}
		)

		m := mockPlayerStorage{
			t: t,
			create: asset.PlayerCreate{
				PlayerChange: asset.PlayerChange{
					Name:        name,
					Description: desc,
					HomeID:      homeID,
					LocationID:  locID,
				},
			},
			createPlayer: &asset.Player{
				ID:          playerID,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locID,
				Created:     created,
				Updated:     updated,
			},
		}

		createReq := rest.PlayerCreateRequest{
			PlayerRequest: rest.PlayerRequest{
				Name:        name,
				Description: desc,
				HomeID:      homeID.String(),
				LocationID:  locID.String(),
			},
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokePlayersEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var playerResp rest.PlayerResponse
		assert.Nil(t, json.Unmarshal(respBody, &playerResp))

		assert.Compare(t, playerResp, rest.PlayerResponse{Player: rest.Player{
			ID:          playerID.String(),
			Name:        name,
			Description: desc,
			HomeID:      homeID.String(),
			LocationID:  locID.String(),
			Created:     created,
			Updated:     updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestPlayerUpdate(t *testing.T) {
	t.Run("playerID failure", func(t *testing.T) {
		m := mockPlayerStorage{}

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, "bad_playerID")

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid player id, not a well formed uuid: 'bad_playerID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockPlayerStorage{}

		playerID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokePlayersEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockPlayerStorage{}

		playerID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("update player req failure", func(t *testing.T) {
		m := mockPlayerStorage{}

		tests := []struct {
			req    rest.PlayerRequest
			status int
			errMsg string
		}{
			{
				req: rest.PlayerRequest{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty player name",
			},
			{
				req: rest.PlayerRequest{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: player name exceeds maximum length",
			},
			{
				req: rest.PlayerRequest{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty player description",
			},
			{
				req: rest.PlayerRequest{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: player description exceeds maximum length",
			},
			{
				req: rest.PlayerRequest{
					Name:        randString(256),
					Description: randString(4096),
					HomeID:      "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid homeID: 'bad owner id'",
			},
			{
				req: rest.PlayerRequest{
					Name:        "name",
					Description: "description",
					HomeID:      uuid.New().String(),
					LocationID:  "bad location id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid locationID: 'bad location id', invalid UUID length: 15",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.PlayerUpdateRequest{PlayerRequest: test.req})
			assert.Nil(t, err)

			playerID := uuid.New()
			route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

			w := invokePlayersEndpoint(t, m, http.MethodPut, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("player manager update failure", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			playerID = asset.PlayerID(uuid.New())
			homeID   = asset.RoomID(uuid.New())
			locID    = asset.RoomID(uuid.New())
		)

		m := mockPlayerStorage{
			t:        t,
			updateID: playerID,
			update: asset.PlayerUpdate{
				PlayerChange: asset.PlayerChange{
					Name:        name,
					Description: desc,
					HomeID:      homeID,
					LocationID:  locID,
				},
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		playerReq := rest.PlayerUpdateRequest{
			PlayerRequest: rest.PlayerRequest{
				Name:        name,
				Description: desc,
				HomeID:      homeID.String(),
				LocationID:  locID.String(),
			},
		}
		body, err := json.Marshal(playerReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			playerID = asset.PlayerID(uuid.New())
			homeID   = asset.RoomID(uuid.New())
			locID    = asset.RoomID(uuid.New())
			created  = asset.Timestamp{Time: time.Now()}
			updated  = asset.Timestamp{Time: time.Now()}
		)

		m := mockPlayerStorage{
			t:        t,
			updateID: playerID,
			update: asset.PlayerUpdate{
				PlayerChange: asset.PlayerChange{
					Name:        name,
					Description: desc,
					HomeID:      homeID,
					LocationID:  locID,
				},
			},
			updatePlayer: &asset.Player{
				ID:          playerID,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locID,
				Created:     created,
				Updated:     updated,
			},
		}

		playerReq := rest.PlayerUpdateRequest{
			PlayerRequest: rest.PlayerRequest{
				Name:        name,
				Description: desc,
				HomeID:      homeID.String(),
				LocationID:  locID.String(),
			},
		}
		body, err := json.Marshal(playerReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var playerResp rest.PlayerResponse
		assert.Nil(t, json.Unmarshal(respBody, &playerResp))

		assert.Compare(t, playerResp, rest.PlayerResponse{Player: rest.Player{
			ID:          playerID.String(),
			Name:        name,
			Description: desc,
			HomeID:      homeID.String(),
			LocationID:  locID.String(),
			Created:     created,
			Updated:     updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestPlayerRemove(t *testing.T) {
	playerID := asset.PlayerID(uuid.New())

	t.Run("playerID failure", func(t *testing.T) {
		m := mockPlayerStorage{}

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, "bad_playerID")

		w := invokePlayersEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid player id, not a well formed uuid: 'bad_playerID'")
	})

	t.Run("player manager remove eailure", func(t *testing.T) {
		m := mockPlayerStorage{
			t:         t,
			removeID:  playerID,
			removeErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockPlayerStorage{
			t:        t,
			removeID: playerID,
		}

		route := fmt.Sprintf("%s/%s", server.V1PlayerRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// helper

func invokePlayersEndpoint(t *testing.T, m mockPlayerStorage, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := server.PlayersService{Storage: m}
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

// mockPlayerStorage

type (
	mockPlayerStorage struct {
		t *testing.T

		filter  asset.PlayerFilter
		list    []*asset.Player
		listErr error

		getID     asset.PlayerID
		getPlayer *asset.Player
		getErr    error

		create       asset.PlayerCreate
		createPlayer *asset.Player
		createErr    error

		updateID     asset.PlayerID
		update       asset.PlayerUpdate
		updatePlayer *asset.Player
		updateErr    error

		removeID  asset.PlayerID
		removeErr error
	}
)

func (m mockPlayerStorage) List(ctx context.Context, filter asset.PlayerFilter) ([]*asset.Player, error) {
	assert.Compare(m.t, filter, m.filter)
	return m.list, m.listErr
}

func (m mockPlayerStorage) Get(ctx context.Context, id asset.PlayerID) (*asset.Player, error) {
	assert.Compare(m.t, id, m.getID)
	return m.getPlayer, m.getErr
}

func (m mockPlayerStorage) Create(ctx context.Context, create asset.PlayerCreate) (*asset.Player, error) {
	assert.Compare(m.t, create, m.create)
	return m.createPlayer, m.createErr
}

func (m mockPlayerStorage) Update(ctx context.Context, id asset.PlayerID, update asset.PlayerUpdate) (*asset.Player, error) {
	assert.Compare(m.t, id, m.updateID)
	assert.Compare(m.t, update, m.update)
	return m.updatePlayer, m.updateErr
}

func (m mockPlayerStorage) Remove(ctx context.Context, id asset.PlayerID) error {
	assert.Compare(m.t, id, m.removeID)
	return m.removeErr
}
