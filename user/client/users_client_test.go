package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"
	"arcadium.dev/core/require"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/asset"
	oapi "arcadium.dev/arcade/internal/user/client"
	"arcadium.dev/arcade/user"
	"arcadium.dev/arcade/user/client"
)

func TestUsersClient_List(t *testing.T) {
	var (
		id       = user.ID(uuid.New())
		login    = "ajones"
		pubkey   = []byte("public key goes here")
		playerID = asset.PlayerID(uuid.New())
		created  = arcade.Timestamp{Time: time.Now()}
		updated  = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name   string
		filter user.Filter
		server *httptest.Server
		verify func(*testing.T, []*user.User, error)
	}{
		{
			name:   "bad request",
			filter: user.Filter{Offset: 20, Limit: 10},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrBadRequest))
			})),
			verify: func(t *testing.T, users []*user.User, err error) {
				assert.Nil(t, users)
				assert.Error(t, err, `bad request: error from users server 'bad request: error goes here'`)
			},
		},
		{
			name:   "internal server error",
			filter: user.Filter{Offset: 20, Limit: 10},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrInternal))
			})),
			verify: func(t *testing.T, users []*user.User, err error) {
				assert.Nil(t, users)
				assert.Error(t, err, `internal server error: error from users server 'internal server error: error goes here'`)
			},
		},
		{
			name:   "not implemented",
			filter: user.Filter{Offset: 20, Limit: 10},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotImplemented))
			})),
			verify: func(t *testing.T, users []*user.User, err error) {
				assert.Nil(t, users)
				assert.Error(t, err, `list users failed: unknown response, status: 501 Not Implemented`)
			},
		},
		{
			name:   "bad user id",
			filter: user.Filter{Offset: 200, Limit: 1000},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				err := json.NewEncoder(w).Encode(oapi.UsersResponse{Users: []oapi.User{
					{
						ID:        "bad user id",
						Login:     login,
						PublicKey: string(pubkey),
						PlayerID:  playerID.String(),
						Created:   created,
						Updated:   updated,
					},
				}})
				assert.Nil(t, err)
			})),
			verify: func(t *testing.T, users []*user.User, err error) {
				assert.Nil(t, users)
				assert.Error(t, err, `users client api failed, bad request: received invalid user ID: 'bad user id': invalid UUID length: 11`)
			},
		},
		{
			name:   "bad player id",
			filter: user.Filter{Offset: 200, Limit: 1000},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				err := json.NewEncoder(w).Encode(oapi.UsersResponse{Users: []oapi.User{
					{
						ID:        id.String(),
						Login:     login,
						PublicKey: string(pubkey),
						PlayerID:  "bad player id",
						Created:   created,
						Updated:   updated,
					},
				}})
				assert.Nil(t, err)
			})),
			verify: func(t *testing.T, users []*user.User, err error) {
				assert.Nil(t, users)
				assert.Error(t, err, `users client api failed, bad request: received invalid user playerID: 'bad player id': invalid UUID length: 13`)
			},
		},
		{
			name:   "success",
			filter: user.Filter{Offset: 200, Limit: 1000},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				err := json.NewEncoder(w).Encode(oapi.UsersResponse{Users: []oapi.User{
					{
						ID:        id.String(),
						Login:     login,
						PublicKey: string(pubkey),
						PlayerID:  playerID.String(),
						Created:   created,
						Updated:   updated,
					},
				}})
				assert.Nil(t, err)
			})),
			verify: func(t *testing.T, users []*user.User, err error) {
				assert.Nil(t, err)
				require.Equal(t, len(users), 1)
				assert.Compare(t, *users[0], user.User{
					ID:        id,
					Login:     login,
					PublicKey: pubkey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			c, err := client.New(test.server.URL)
			require.Nil(t, err)
			users, err := c.List(context.Background(), test.filter)
			test.verify(t, users, err)
		})
	}
}

func TestUsersClient_Get(t *testing.T) {
	var (
		id       = user.ID(uuid.New())
		login    = "ajones"
		pubkey   = []byte("public key goes here")
		playerID = asset.PlayerID(uuid.New())
		created  = arcade.Timestamp{Time: time.Now()}
		updated  = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name   string
		server *httptest.Server
		verify func(*testing.T, *user.User, error)
	}{
		{
			name: "bad request",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrBadRequest))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `bad request: error from users server 'bad request: error goes here'`)
			},
		},
		{
			name: "not found",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotFound))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `not found: error from users server 'not found: error goes here'`)
			},
		},
		{
			name: "internal server error",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrInternal))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `internal server error: error from users server 'internal server error: error goes here'`)
			},
		},
		{
			name: "not implemented",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotImplemented))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `get user failed: unknown response, status: 501 Not Implemented`)
			},
		},
		{
			name: "success",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				err := json.NewEncoder(w).Encode(oapi.UserResponse{User: oapi.User{
					ID:        id.String(),
					Login:     login,
					PublicKey: string(pubkey),
					PlayerID:  playerID.String(),
					Created:   created,
					Updated:   updated,
				}})
				assert.Nil(t, err)
			})),
			verify: func(t *testing.T, u *user.User, err error) {
				assert.Nil(t, err)
				require.NotNil(t, u)
				assert.Compare(t, *u, user.User{
					ID:        id,
					Login:     login,
					PublicKey: pubkey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			c, err := client.New(test.server.URL)
			require.Nil(t, err)
			user, err := c.Get(context.Background(), id)
			test.verify(t, user, err)
		})
	}
}

func TestUsersClient_Create(t *testing.T) {
	var (
		id       = user.ID(uuid.New())
		login    = "ajones"
		pubkey   = []byte("public key goes here")
		playerID = asset.PlayerID(uuid.New())
		created  = arcade.Timestamp{Time: time.Now()}
		updated  = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name   string
		server *httptest.Server
		verify func(*testing.T, *user.User, error)
	}{
		{
			name: "bad request",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrBadRequest))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `bad request: error from users server 'bad request: error goes here'`)
			},
		},
		{
			name: "conflict",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrConflict))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `conflict: error from users server 'conflict: error goes here'`)
			},
		},
		{
			name: "internal server error",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrInternal))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `internal server error: error from users server 'internal server error: error goes here'`)
			},
		},
		{
			name: "not implemented",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotImplemented))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `create user failed: unknown response, status: 501 Not Implemented`)
			},
		},
		{
			name: "success",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.WriteHeader(http.StatusCreated)
				err := json.NewEncoder(w).Encode(oapi.UserResponse{User: oapi.User{
					ID:        id.String(),
					Login:     login,
					PublicKey: string(pubkey),
					PlayerID:  playerID.String(),
					Created:   created,
					Updated:   updated,
				}})
				assert.Nil(t, err)
			})),
			verify: func(t *testing.T, u *user.User, err error) {
				assert.Nil(t, err)
				require.NotNil(t, u)
				assert.Compare(t, *u, user.User{
					ID:        id,
					Login:     login,
					PublicKey: pubkey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			c, err := client.New(test.server.URL)
			require.Nil(t, err)
			user, err := c.Create(context.Background(), user.Create{Change: user.Change{Login: login, PublicKey: pubkey}})
			test.verify(t, user, err)
		})
	}
}

func TestUsersClient_Update(t *testing.T) {
	var (
		id       = user.ID(uuid.New())
		login    = "ajones"
		pubkey   = []byte("public key goes here")
		playerID = asset.PlayerID(uuid.New())
		created  = arcade.Timestamp{Time: time.Now()}
		updated  = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name   string
		server *httptest.Server
		verify func(*testing.T, *user.User, error)
	}{
		{
			name: "bad request",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrBadRequest))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `bad request: error from users server 'bad request: error goes here'`)
			},
		},
		{
			name: "not found",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotFound))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `not found: error from users server 'not found: error goes here'`)
			},
		},
		{
			name: "internal server error",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrInternal))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `internal server error: error from users server 'internal server error: error goes here'`)
			},
		},
		{
			name: "not implemented",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotImplemented))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `update user failed: unknown response, status: 501 Not Implemented`)
			},
		},
		{
			name: "success",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				err := json.NewEncoder(w).Encode(oapi.UserResponse{User: oapi.User{
					ID:        id.String(),
					Login:     login,
					PublicKey: string(pubkey),
					PlayerID:  playerID.String(),
					Created:   created,
					Updated:   updated,
				}})
				assert.Nil(t, err)
			})),
			verify: func(t *testing.T, u *user.User, err error) {
				assert.Nil(t, err)
				require.NotNil(t, u)
				assert.Compare(t, *u, user.User{
					ID:        id,
					Login:     login,
					PublicKey: pubkey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			c, err := client.New(test.server.URL)
			require.Nil(t, err)
			user, err := c.Update(context.Background(), id, user.Update{Change: user.Change{Login: login, PublicKey: pubkey}})
			test.verify(t, user, err)
		})
	}
}

func TestUsersClient_AssociatePlayer(t *testing.T) {
	var (
		id       = user.ID(uuid.New())
		login    = "ajones"
		pubkey   = []byte("public key goes here")
		playerID = asset.PlayerID(uuid.New())
		created  = arcade.Timestamp{Time: time.Now()}
		updated  = arcade.Timestamp{Time: time.Now()}
	)

	tests := []struct {
		name   string
		server *httptest.Server
		verify func(*testing.T, *user.User, error)
	}{
		{
			name: "bad request",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrBadRequest))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `bad request: error from users server 'bad request: error goes here'`)
			},
		},
		{
			name: "not found",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotFound))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `not found: error from users server 'not found: error goes here'`)
			},
		},
		{
			name: "internal server error",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrInternal))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `internal server error: error from users server 'internal server error: error goes here'`)
			},
		},
		{
			name: "not implemented",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotImplemented))
			})),
			verify: func(t *testing.T, user *user.User, err error) {
				assert.Nil(t, user)
				assert.Error(t, err, `associate player with user failed: unknown response, status: 501 Not Implemented`)
			},
		},
		{
			name: "success",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				err := json.NewEncoder(w).Encode(oapi.UserResponse{User: oapi.User{
					ID:        id.String(),
					Login:     login,
					PublicKey: string(pubkey),
					PlayerID:  playerID.String(),
					Created:   created,
					Updated:   updated,
				}})
				assert.Nil(t, err)
			})),
			verify: func(t *testing.T, u *user.User, err error) {
				assert.Nil(t, err)
				require.NotNil(t, u)
				assert.Compare(t, *u, user.User{
					ID:        id,
					Login:     login,
					PublicKey: pubkey,
					PlayerID:  playerID,
					Created:   created,
					Updated:   updated,
				}, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			c, err := client.New(test.server.URL)
			require.Nil(t, err)
			user, err := c.AssociatePlayer(context.Background(), id, user.AssociatePlayer{PlayerID: playerID})
			test.verify(t, user, err)
		})
	}
}

func TestUsersClient_Remove(t *testing.T) {
	var (
		id = user.ID(uuid.New())
	)

	tests := []struct {
		name   string
		id     user.ID
		server *httptest.Server
		verify func(*testing.T, error)
	}{
		{
			name: "internal server error",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrInternal))
			})),
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, `internal server error: error from users server 'internal server error: error goes here'`)
			},
		},
		{
			name: "not implemented",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.Response(r.Context(), w, fmt.Errorf("%w: error goes here", errors.ErrNotImplemented))
			})),
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, `remove user failed: unknown response, status: 501 Not Implemented`)
			},
		},
		{
			name: "success",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			})),
			verify: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			c, err := client.New(test.server.URL)
			require.Nil(t, err)
			err = c.Remove(context.Background(), id)
			test.verify(t, err)
		})
	}
}
