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
	"arcadium.dev/arcade/user"
	"arcadium.dev/arcade/user/rest"
	"arcadium.dev/arcade/user/rest/server"
)

func TestUsersList(t *testing.T) {
	route := server.V1UserRoute

	t.Run("new filter failure", func(t *testing.T) {
		m := mockUserStorage{}

		w := invokeUsersEndpoint(t, m, http.MethodGet, route, nil, "offset", "bad offset")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid offset query parameter: 'bad offset'")

		w = invokeUsersEndpoint(t, m, http.MethodGet, route, nil, "limit", "bad limit")
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid limit query parameter: 'bad limit'")
	})

	t.Run("user manager list failure", func(t *testing.T) {
		m := mockUserStorage{
			t: t,
			filter: user.Filter{
				Offset: 10,
				Limit:  10,
			},
			listErr: fmt.Errorf("%w: list failure", errors.ErrNotFound),
		}

		w := invokeUsersEndpoint(t, m, http.MethodGet, route, nil, "offset", "10", "limit", "10")
		assertRespError(t, w, http.StatusNotFound, "list failure")
	})

	t.Run("success", func(t *testing.T) {
		var (
			userID    = user.ID(uuid.New())
			login     = "ajones"
			publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
			playerID  = asset.PlayerID(uuid.New())
			created   = arcade.Timestamp{Time: time.Now()}
			updated   = arcade.Timestamp{Time: time.Now()}
		)

		m := mockUserStorage{
			t: t,
			filter: user.Filter{
				Offset: 25,
				Limit:  100,
			},
			list: []*user.User{
				{
					ID:        userID,
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				},
			},
		}

		w := invokeUsersEndpoint(t, m, http.MethodGet, route, nil, "offset", "25", "limit", "100")

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var usersResp rest.UsersResponse
		assert.Nil(t, json.Unmarshal(body, &usersResp))

		assert.Compare(t, usersResp, rest.UsersResponse{Users: []rest.User{
			{
				ID:        userID.String(),
				Login:     login,
				PublicKey: string(publicKey),
				PlayerID:  playerID.String(),
				Created:   created,
				Updated:   updated,
			},
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestUserGet(t *testing.T) {
	userID := user.ID(uuid.New())

	t.Run("userID failure", func(t *testing.T) {
		m := mockUserStorage{}

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, "bad_userID")

		w := invokeUsersEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid user id, not a well formed uuid: 'bad_userID'")
	})

	t.Run("user manager get failure", func(t *testing.T) {
		m := mockUserStorage{
			t:      t,
			getID:  userID,
			getErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

		w := invokeUsersEndpoint(t, m, http.MethodGet, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		var (
			login     = "ajones"
			publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
			playerID  = asset.PlayerID(uuid.New())
			created   = arcade.Timestamp{Time: time.Now()}
			updated   = arcade.Timestamp{Time: time.Now()}
		)

		m := mockUserStorage{
			t:     t,
			getID: userID,
			getUser: &user.User{
				ID:        userID,
				Login:     login,
				PublicKey: publicKey,
				PlayerID:  playerID,
				Created:   created,
				Updated:   updated,
			},
		}

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

		w := invokeUsersEndpoint(t, m, http.MethodGet, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var userResp rest.UserResponse
		assert.Nil(t, json.Unmarshal(body, &userResp))

		assert.Compare(t, userResp, rest.UserResponse{User: rest.User{
			ID:        userID.String(),
			Login:     login,
			PublicKey: string(publicKey),
			PlayerID:  playerID.String(),
			Created:   created,
			Updated:   updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestUserCreate(t *testing.T) {
	var (
		login = "ajones"
	)

	route := server.V1UserRoute

	t.Run("empty body", func(t *testing.T) {
		m := mockUserStorage{}

		w := invokeUsersEndpoint(t, m, http.MethodPost, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeUsersEndpoint(t, m, http.MethodPost, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockUserStorage{}

		w := invokeUsersEndpoint(t, m, http.MethodPost, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("create user req failure", func(t *testing.T) {
		m := mockUserStorage{}

		tests := []struct {
			req    rest.UserRequest
			status int
			errMsg string
		}{
			{
				req: rest.UserRequest{
					Login: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty user login",
			},
			{
				req: rest.UserRequest{
					Login: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: user login exceeds maximum length",
			},
			{
				req: rest.UserRequest{
					Login:     login,
					PublicKey: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty user ssh public key",
			},
			{
				req: rest.UserRequest{
					Login:     login,
					PublicKey: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: user ssh public key exceeds maximum length",
			},
			{
				req: rest.UserRequest{
					Login:     randString(256),
					PublicKey: randString(4096),
					PlayerID:  "bad player id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid playerID: 'bad player id'",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.UserCreateRequest{UserRequest: test.req})
			assert.Nil(t, err)

			w := invokeUsersEndpoint(t, m, http.MethodPost, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("user manager create failure", func(t *testing.T) {
		var (
			login     = "ajones"
			publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
			playerID  = asset.PlayerID(uuid.New())
		)

		m := mockUserStorage{
			t: t,
			create: user.Create{
				Change: user.Change{
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
				},
			},
			createErr: fmt.Errorf("%w: %s", errors.ErrConflict, "create failure"),
		}

		create := rest.UserCreateRequest{
			UserRequest: rest.UserRequest{
				Login:     login,
				PublicKey: string(publicKey),
				PlayerID:  playerID.String(),
			},
		}
		body, err := json.Marshal(create)
		assert.Nil(t, err)

		w := invokeUsersEndpoint(t, m, http.MethodPost, route, body)
		assertRespError(t, w, http.StatusConflict, "conflict: create failure")
	})

	t.Run("success", func(t *testing.T) {
		var (
			userID    = user.ID(uuid.New())
			login     = "ajones"
			publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
			playerID  = asset.PlayerID(uuid.New())
			created   = arcade.Timestamp{Time: time.Now()}
			updated   = arcade.Timestamp{Time: time.Now()}
		)

		m := mockUserStorage{
			t: t,
			create: user.Create{
				Change: user.Change{
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
				},
			},
			createUser: &user.User{
				ID:        userID,
				Login:     login,
				PublicKey: publicKey,
				PlayerID:  playerID,
				Created:   created,
				Updated:   updated,
			},
		}

		createReq := rest.UserCreateRequest{
			UserRequest: rest.UserRequest{
				Login:     login,
				PublicKey: string(publicKey),
				PlayerID:  playerID.String(),
			},
		}
		body, err := json.Marshal(createReq)
		assert.Nil(t, err)

		w := invokeUsersEndpoint(t, m, http.MethodPost, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusCreated)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var userResp rest.UserResponse
		assert.Nil(t, json.Unmarshal(respBody, &userResp))

		assert.Compare(t, userResp, rest.UserResponse{User: rest.User{
			ID:        userID.String(),
			Login:     login,
			PublicKey: string(publicKey),
			PlayerID:  playerID.String(),
			Created:   created,
			Updated:   updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestUserUpdate(t *testing.T) {
	t.Run("userID failure", func(t *testing.T) {
		m := mockUserStorage{}

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, "bad_userID")

		w := invokeUsersEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid user id, not a well formed uuid: 'bad_userID'")
	})

	t.Run("empty body", func(t *testing.T) {
		m := mockUserStorage{}

		userID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

		w := invokeUsersEndpoint(t, m, http.MethodPut, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")

		w = invokeUsersEndpoint(t, m, http.MethodPut, route, []byte(""))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid json: a json encoded body is required")
	})

	t.Run("invalid body", func(t *testing.T) {
		m := mockUserStorage{}

		userID := uuid.New()
		route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

		w := invokeUsersEndpoint(t, m, http.MethodPut, route, []byte(`{"id": `))
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid body: unexpected end of JSON input")
	})

	t.Run("update user req failure", func(t *testing.T) {
		var (
			login = "ajones"
		)

		m := mockUserStorage{}

		tests := []struct {
			req    rest.UserRequest
			status int
			errMsg string
		}{
			{
				req: rest.UserRequest{
					Login: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty user login",
			},
			{
				req: rest.UserRequest{
					Login: randString(257),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: user login exceeds maximum length",
			},
			{
				req: rest.UserRequest{
					Login:     login,
					PublicKey: "",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: empty user ssh public key",
			},
			{
				req: rest.UserRequest{
					Login:     login,
					PublicKey: randString(4097),
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: user ssh public key exceeds maximum length",
			},
			{
				req: rest.UserRequest{
					Login:     randString(256),
					PublicKey: randString(4096),
					PlayerID:  "bad player id",
				},
				status: http.StatusBadRequest,
				errMsg: "bad request: invalid playerID: 'bad player id'",
			},
		}

		for _, test := range tests {
			body, err := json.Marshal(rest.UserUpdateRequest{UserRequest: test.req})
			assert.Nil(t, err)

			userID := uuid.New()
			route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

			w := invokeUsersEndpoint(t, m, http.MethodPut, route, body)
			assertRespError(t, w, test.status, test.errMsg)
		}
	})

	t.Run("user manager update failure", func(t *testing.T) {
		var (
			userID    = user.ID(uuid.New())
			login     = "ajones"
			publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
			playerID  = asset.PlayerID(uuid.New())
		)

		m := mockUserStorage{
			t:        t,
			updateID: userID,
			update: user.Update{
				Change: user.Change{
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
				},
			},
			updateErr: fmt.Errorf("%w: %s", errors.ErrNotFound, "update failure"),
		}

		updateReq := rest.UserUpdateRequest{
			UserRequest: rest.UserRequest{
				Login:     login,
				PublicKey: string(publicKey),
				PlayerID:  playerID.String(),
			},
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

		w := invokeUsersEndpoint(t, m, http.MethodPut, route, body)
		assertRespError(t, w, http.StatusNotFound, "not found: update failure")
	})

	t.Run("success", func(t *testing.T) {
		var (
			userID    = user.ID(uuid.New())
			login     = "ajones"
			publicKey = []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHXPLG0x/V6kk7BdlgY1YR61xWjt3HLvEhdlscUs4GjO foo@bar")
			playerID  = asset.PlayerID(uuid.New())
			created   = arcade.Timestamp{Time: time.Now()}
			updated   = arcade.Timestamp{Time: time.Now()}
		)

		m := mockUserStorage{
			t:        t,
			updateID: userID,
			update: user.Update{
				Change: user.Change{
					Login:     login,
					PublicKey: publicKey,
					PlayerID:  playerID,
				},
			},
			updateUser: &user.User{
				ID:        userID,
				Login:     login,
				PublicKey: publicKey,
				PlayerID:  playerID,
				Created:   created,
				Updated:   updated,
			},
		}

		updateReq := rest.UserUpdateRequest{
			UserRequest: rest.UserRequest{
				Login:     login,
				PublicKey: string(publicKey),
				PlayerID:  playerID.String(),
			},
		}
		body, err := json.Marshal(updateReq)
		assert.Nil(t, err)

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

		w := invokeUsersEndpoint(t, m, http.MethodPut, route, body)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		respBody, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		defer resp.Body.Close()

		var userResp rest.UserResponse
		assert.Nil(t, json.Unmarshal(respBody, &userResp))

		assert.Compare(t, userResp, rest.UserResponse{User: rest.User{
			ID:        userID.String(),
			Login:     login,
			PublicKey: string(publicKey),
			PlayerID:  playerID.String(),
			Created:   created,
			Updated:   updated,
		}}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestUserRemove(t *testing.T) {
	userID := user.ID(uuid.New())

	t.Run("userID failure", func(t *testing.T) {
		m := mockUserStorage{}

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, "bad_userID")

		w := invokeUsersEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: invalid user id, not a well formed uuid: 'bad_userID'")
	})

	t.Run("user manager remove failure", func(t *testing.T) {
		m := mockUserStorage{
			t:         t,
			removeID:  userID,
			removeErr: fmt.Errorf("%w: %s", errors.ErrBadRequest, "get failure"),
		}

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

		w := invokeUsersEndpoint(t, m, http.MethodDelete, route, nil)
		assertRespError(t, w, http.StatusBadRequest, "bad request: get failure")
	})

	t.Run("success", func(t *testing.T) {
		m := mockUserStorage{
			t:        t,
			removeID: userID,
		}

		route := fmt.Sprintf("%s/%s", server.V1UserRoute, userID.String())

		w := invokeUsersEndpoint(t, m, http.MethodDelete, route, nil)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})
}

// helper

func invokeUsersEndpoint(t *testing.T, m mockUserStorage, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := server.UsersService{Storage: m}
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

// mockUserStorage

type (
	mockUserStorage struct {
		t *testing.T

		filter  user.Filter
		list    []*user.User
		listErr error

		getID   user.ID
		getUser *user.User
		getErr  error

		create     user.Create
		createUser *user.User
		createErr  error

		updateID   user.ID
		update     user.Update
		updateUser *user.User
		updateErr  error

		removeID  user.ID
		removeErr error
	}
)

func (m mockUserStorage) List(ctx context.Context, filter user.Filter) ([]*user.User, error) {
	assert.Compare(m.t, filter, m.filter)
	return m.list, m.listErr
}

func (m mockUserStorage) Get(ctx context.Context, id user.ID) (*user.User, error) {
	assert.Compare(m.t, id, m.getID)
	return m.getUser, m.getErr
}

func (m mockUserStorage) Create(ctx context.Context, create user.Create) (*user.User, error) {
	assert.Compare(m.t, create, m.create)
	return m.createUser, m.createErr
}

func (m mockUserStorage) Update(ctx context.Context, id user.ID, update user.Update) (*user.User, error) {
	assert.Compare(m.t, id, m.updateID)
	assert.Compare(m.t, update, m.update)
	return m.updateUser, m.updateErr
}

func (m mockUserStorage) Remove(ctx context.Context, id user.ID) error {
	assert.Compare(m.t, id, m.removeID)
	return m.removeErr
}
