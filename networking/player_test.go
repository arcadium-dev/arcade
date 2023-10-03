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

func TestPlayersList(t *testing.T) {
	route := networking.V1PlayersRoute
	id := uuid.New()

	t.Run("new filter failure", func(t *testing.T) {
		m := mockPlayerManager{}

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
		m := mockPlayerManager{
			t: t,
			filter: arcade.PlayersFilter{
				LocationID: arcade.RoomID(id),
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
			playerID   = arcade.PlayerID(uuid.New())
			homeID     = arcade.RoomID(uuid.New())
			locationID = arcade.RoomID(uuid.New())
			created    = arcade.Timestamp{Time: time.Now()}
			updated    = arcade.Timestamp{Time: time.Now()}
		)

		m := mockPlayerManager{
			t: t,
			filter: arcade.PlayersFilter{
				Offset: 25,
				Limit:  100,
			},
			list: []*arcade.Player{
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

		var egressPlayers networking.EgressPlayers
		assert.Nil(t, json.Unmarshal(body, &egressPlayers))

		assert.Compare(t, egressPlayers, networking.EgressPlayers{Players: []networking.Player{
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
	playerID := arcade.PlayerID(uuid.New())

	t.Run("playerID failure", func(t *testing.T) {
		m := mockPlayerManager{}

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, "bad_playerID")

		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid playerID, not a well formed uuid: 'bad_playerID'")
	})

	t.Run("player manager get failure", func(t *testing.T) {
		m := mockPlayerManager{
			t:           t,
			getPlayerID: playerID,
			getErr:      fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "desc"
		)
		var (
			homeID     = arcade.RoomID(uuid.New())
			locationID = arcade.RoomID(homeID)
			created    = arcade.Timestamp{Time: time.Now()}
			updated    = arcade.Timestamp{Time: time.Now()}
		)

		m := mockPlayerManager{
			t:           t,
			getPlayerID: playerID,
			getPlayer: &arcade.Player{
				ID:          playerID,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			},
		}

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var egressPlayer networking.EgressPlayer
		assert.Nil(t, json.Unmarshal(body, &egressPlayer))

		assert.Compare(t, egressPlayer, networking.EgressPlayer{Player: networking.Player{
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
	route := networking.V1PlayersRoute

	t.Run("empty body", func(t *testing.T) {
		m := mockPlayerManager{}

		w := invokePlayersEndpoint(t, m, http.MethodPost, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokePlayersEndpoint(t, m, http.MethodPost, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockPlayerManager{}

		w := invokePlayersEndpoint(t, m, http.MethodPost, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("ingress player req failure", func(t *testing.T) {
		m := mockPlayerManager{}

		tests := []struct {
			ingressPlayer networking.IngressPlayer
			status        int
			errMsg        string
		}{
			{
				ingressPlayer: networking.IngressPlayer{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty player name",
			},
			{
				ingressPlayer: networking.IngressPlayer{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: player name exceeds maximum length",
			},
			{
				ingressPlayer: networking.IngressPlayer{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty player description",
			},
			{
				ingressPlayer: networking.IngressPlayer{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: player description exceeds maximum length",
			},
			{
				ingressPlayer: networking.IngressPlayer{
					Name:        randString(256),
					Description: randString(4096),
					HomeID:      "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid homeID: 'bad owner id'",
			},
			{
				ingressPlayer: networking.IngressPlayer{
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
			body, err := json.Marshal(test.ingressPlayer)
			assert.Nil(t, err)

			w := invokePlayersEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("player manager create failure", func(t *testing.T) {
		var (
			homeID = arcade.RoomID(uuid.New())
			locID  = arcade.RoomID(uuid.New())
		)

		m := mockPlayerManager{
			t: t,
			createPlayerReq: arcade.IngressPlayer{
				Name:        "name",
				Description: "description",
				HomeID:      homeID,
				LocationID:  locID,
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		ingressPlayer := networking.IngressPlayer{
			Name:        "name",
			Description: "description",
			HomeID:      homeID.String(),
			LocationID:  locID.String(),
		}
		body, err := json.Marshal(ingressPlayer)
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
			playerID = arcade.PlayerID(uuid.New())
			homeID   = arcade.RoomID(uuid.New())
			locID    = arcade.RoomID(uuid.New())
			created  = arcade.Timestamp{Time: time.Now()}
			updated  = arcade.Timestamp{Time: time.Now()}
		)

		m := mockPlayerManager{
			t: t,
			createPlayerReq: arcade.IngressPlayer{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locID,
			},
			createPlayer: &arcade.Player{
				ID:          playerID,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locID,
				Created:     created,
				Updated:     updated,
			},
		}

		ingressPlayer := networking.IngressPlayer{
			Name:        name,
			Description: desc,
			HomeID:      homeID.String(),
			LocationID:  locID.String(),
		}
		body, err := json.Marshal(ingressPlayer)
		assert.Nil(t, err)

		w := invokePlayersEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var playerResp networking.EgressPlayer
		assert.Nil(t, json.Unmarshal(respBody, &playerResp))

		assert.Compare(t, playerResp, networking.EgressPlayer{Player: networking.Player{
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
		m := mockPlayerManager{}

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, "bad_playerID")

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid playerID, not a well formed uuid: 'bad_playerID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockPlayerManager{}

		playerID := uuid.New()
		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokePlayersEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockPlayerManager{}

		playerID := uuid.New()
		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("ingress player req failure", func(t *testing.T) {
		m := mockPlayerManager{}

		tests := []struct {
			ingressPlayer networking.IngressPlayer
			status        int
			errMsg        string
		}{
			{
				ingressPlayer: networking.IngressPlayer{
					Name: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty player name",
			},
			{
				ingressPlayer: networking.IngressPlayer{
					Name: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: player name exceeds maximum length",
			},
			{
				ingressPlayer: networking.IngressPlayer{
					Name:        "name",
					Description: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty player description",
			},
			{
				ingressPlayer: networking.IngressPlayer{
					Name:        "name",
					Description: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: player description exceeds maximum length",
			},
			{
				ingressPlayer: networking.IngressPlayer{
					Name:        randString(256),
					Description: randString(4096),
					HomeID:      "bad owner id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid homeID: 'bad owner id'",
			},
			{
				ingressPlayer: networking.IngressPlayer{
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
			body, err := json.Marshal(test.ingressPlayer)
			assert.Nil(t, err)

			playerID := uuid.New()
			route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

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
			playerID = arcade.PlayerID(uuid.New())
			homeID   = arcade.RoomID(uuid.New())
			locID    = arcade.RoomID(uuid.New())
		)

		m := mockPlayerManager{
			t:              t,
			updatePlayerID: playerID,
			updatePlayerReq: arcade.IngressPlayer{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locID,
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		ingressPlayer := networking.IngressPlayer{
			Name:        name,
			Description: desc,
			HomeID:      homeID.String(),
			LocationID:  locID.String(),
		}
		body, err := json.Marshal(ingressPlayer)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		const (
			name = "name"
			desc = "description"
		)
		var (
			playerID = arcade.PlayerID(uuid.New())
			homeID   = arcade.RoomID(uuid.New())
			locID    = arcade.RoomID(uuid.New())
			created  = arcade.Timestamp{Time: time.Now()}
			updated  = arcade.Timestamp{Time: time.Now()}
		)

		m := mockPlayerManager{
			t:              t,
			updatePlayerID: playerID,
			updatePlayerReq: arcade.IngressPlayer{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locID,
			},
			updatePlayer: &arcade.Player{
				ID:          playerID,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locID,
				Created:     created,
				Updated:     updated,
			},
		}

		ingressPlayer := networking.IngressPlayer{
			Name:        name,
			Description: desc,
			HomeID:      homeID.String(),
			LocationID:  locID.String(),
		}
		body, err := json.Marshal(ingressPlayer)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var playerResp networking.EgressPlayer
		assert.Nil(t, json.Unmarshal(respBody, &playerResp))

		assert.Compare(t, playerResp, networking.EgressPlayer{Player: networking.Player{
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
	playerID := arcade.PlayerID(uuid.New())

	t.Run("playerID failure", func(t *testing.T) {
		m := mockPlayerManager{}

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, "bad_playerID")

		w := invokePlayersEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid playerID, not a well formed uuid: 'bad_playerID'")
	})

	t.Run("player manager remove eailure", func(t *testing.T) {
		m := mockPlayerManager{
			t:              t,
			removePlayerID: playerID,
			removeErr:      fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockPlayerManager{
			t:              t,
			removePlayerID: playerID,
		}

		route := fmt.Sprintf("%s/%s", networking.V1PlayersRoute, playerID.String())

		w := invokePlayersEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// mockPlayerManager

type (
	mockPlayerManager struct {
		t *testing.T

		filter  arcade.PlayersFilter
		list    []*arcade.Player
		listErr error

		getPlayerID arcade.PlayerID
		getPlayer   *arcade.Player
		getErr      error

		createPlayerReq arcade.IngressPlayer
		createPlayer    *arcade.Player
		createErr       error

		updatePlayerID  arcade.PlayerID
		updatePlayerReq arcade.IngressPlayer
		updatePlayer    *arcade.Player
		updateErr       error

		removePlayerID arcade.PlayerID
		removeErr      error
	}
)

func (m mockPlayerManager) List(ctx context.Context, filter arcade.PlayersFilter) ([]*arcade.Player, error) {
	assert.Compare(m.t, filter, m.filter)
	return m.list, m.listErr
}

func (m mockPlayerManager) Get(ctx context.Context, playerID arcade.PlayerID) (*arcade.Player, error) {
	assert.Compare(m.t, playerID, m.getPlayerID)
	return m.getPlayer, m.getErr
}

func (m mockPlayerManager) Create(ctx context.Context, ingressPlayer arcade.IngressPlayer) (*arcade.Player, error) {
	assert.Compare(m.t, ingressPlayer, m.createPlayerReq)
	return m.createPlayer, m.createErr
}

func (m mockPlayerManager) Update(ctx context.Context, playerID arcade.PlayerID, ingressPlayer arcade.IngressPlayer) (*arcade.Player, error) {
	assert.Compare(m.t, playerID, m.updatePlayerID)
	assert.Compare(m.t, ingressPlayer, m.updatePlayerReq)
	return m.updatePlayer, m.updateErr
}

func (m mockPlayerManager) Remove(ctx context.Context, playerID arcade.PlayerID) error {
	assert.Compare(m.t, playerID, m.removePlayerID)
	return m.removeErr
}
