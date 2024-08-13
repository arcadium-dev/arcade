package client_test

import (
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

	"arcadium.dev/arcade"
	"arcadium.dev/core/assert"
	"arcadium.dev/core/require"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
	"arcadium.dev/arcade/asset/rest/client"
)

func TestListPlayers(t *testing.T) {
	ctx := context.Background()

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.ListPlayers(ctx, asset.PlayerFilter{})

		assert.Contains(t, err.Error(), `failed to list players: parse "1234:bad url/v1/player": first path segment in URL cannot contain colon`)
	})

	t.Run("max limit failure", func(t *testing.T) {
		c := client.New("https://example.com")

		_, err := c.ListPlayers(ctx, asset.PlayerFilter{Limit: 1000})

		assert.Contains(t, err.Error(), `failed to list players: player filter limit 1000 exceeds maximum 100`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListPlayers(ctx, asset.PlayerFilter{})

		assert.Error(t, err, `failed to list players: 500, Internal Server Error`)
	})

	t.Run("response decode failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ foo`)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListPlayers(ctx, asset.PlayerFilter{})

		assert.Error(t, err, `failed to list players: invalid character 'f' looking for beginning of object key string`)
	})

	t.Run("translate player failure", func(t *testing.T) {
		players := []rest.Player{
			{ID: "bad uuid"},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.PlayersResponse{Players: players})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListPlayers(ctx, asset.PlayerFilter{})

		assert.Error(t, err, `failed to list players: received invalid player ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id   = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			home = "1f290a67-ef2c-455b-aad7-6a3e72276ab5"
			loc  = "8f314204-49f5-44c0-83f1-1e3a24eec3ad"
			name = "name"
			desc = "desc"
		)
		var (
			playerID   = asset.PlayerID(uuid.MustParse(id))
			homeID     = asset.RoomID(uuid.MustParse(home))
			locationID = asset.RoomID(uuid.MustParse(loc))
			created    = arcade.Timestamp{Time: time.Now().UTC()}
			updated    = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rPlayers := []rest.Player{
			{
				ID:          id,
				Name:        name,
				Description: desc,
				HomeID:      home,
				LocationID:  loc,
				Created:     created,
				Updated:     updated,
			},
		}

		aPlayers := []*asset.Player{
			{
				ID:          playerID,
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
				Created:     created,
				Updated:     updated,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			require.Equal(t, len(q["locationID"]), 1)
			assert.Equal(t, q["locationID"][0], loc)
			require.Equal(t, len(q["offset"]), 1)
			assert.Equal(t, q["offset"][0], "10")
			require.Equal(t, len(q["limit"]), 1)
			assert.Equal(t, q["limit"][0], "10")

			err := json.NewEncoder(w).Encode(rest.PlayersResponse{Players: rPlayers})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		filter := asset.PlayerFilter{
			LocationID: locationID,
			Offset:     10,
			Limit:      10,
		}

		players, err := c.ListPlayers(ctx, filter)

		assert.Nil(t, err)
		assert.Equal(t, len(players), 1)
		assert.Compare(t, *players[0], *aPlayers[0], cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestGetPlayer(t *testing.T) {
	var (
		ctx = context.Background()
		id  = asset.PlayerID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.GetPlayer(ctx, id)

		assert.Contains(t, err.Error(), `failed to get player: parse "1234:bad url/v1/player/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetPlayer(ctx, id)

		assert.Error(t, err, `failed to get player: 500, Internal Server Error`)
	})

	t.Run("response decode failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ foo`)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetPlayer(ctx, id)

		assert.Error(t, err, `failed to get player: invalid character 'f' looking for beginning of object key string`)
	})

	t.Run("translate player failure", func(t *testing.T) {
		player := rest.Player{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.PlayerResponse{Player: player})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetPlayer(ctx, id)

		assert.Error(t, err, `failed to get player: received invalid player ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id   = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			home = "1f290a67-ef2c-455b-aad7-6a3e72276ab5"
			loc  = "8f314204-49f5-44c0-83f1-1e3a24eec3ad"
			name = "name"
			desc = "desc"
		)
		var (
			playerID   = asset.PlayerID(uuid.MustParse(id))
			homeID     = asset.RoomID(uuid.MustParse(home))
			locationID = asset.RoomID(uuid.MustParse(loc))
			created    = arcade.Timestamp{Time: time.Now().UTC()}
			updated    = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rPlayer := rest.Player{
			ID:          id,
			Name:        name,
			Description: desc,
			HomeID:      home,
			LocationID:  loc,
			Created:     created,
			Updated:     updated,
		}

		aPlayer := &asset.Player{
			ID:          playerID,
			Name:        name,
			Description: desc,
			HomeID:      homeID,
			LocationID:  locationID,
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.PlayerResponse{Player: rPlayer})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		player, err := c.GetPlayer(ctx, playerID)

		assert.Nil(t, err)
		assert.Compare(t, player, aPlayer, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestCreatePlayer(t *testing.T) {
	const (
		name = "name"
		desc = "desc"
		home = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		loc  = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
	)

	var (
		ctx        = context.Background()
		homeID     = asset.RoomID(uuid.MustParse(home))
		locationID = asset.RoomID(uuid.MustParse(loc))
	)

	t.Run("player change failure", func(t *testing.T) {
		c := client.Client{}

		_, err := c.CreatePlayer(ctx, asset.PlayerCreate{PlayerChange: asset.PlayerChange{Name: ""}})

		assert.Error(t, err, `failed to create player: attempted to send empty name in request`)
	})

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
			},
		})

		assert.Contains(t, err.Error(), `failed to create player: parse "1234:bad url/v1/player": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
			},
		})

		assert.Error(t, err, `failed to create player: 500, Internal Server Error`)
	})

	t.Run("translate player failure", func(t *testing.T) {
		player := rest.Player{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.PlayerResponse{Player: player})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
			},
		})

		assert.Error(t, err, `failed to create player: received invalid player ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		)
		var (
			playerID = asset.PlayerID(uuid.MustParse(id))
			created  = arcade.Timestamp{Time: time.Now().UTC()}
			updated  = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rPlayer := rest.Player{
			ID:          id,
			Name:        name,
			Description: desc,
			HomeID:      home,
			LocationID:  loc,
			Created:     created,
			Updated:     updated,
		}

		aPlayer := &asset.Player{
			ID:          playerID,
			Name:        name,
			Description: desc,
			HomeID:      asset.RoomID(homeID),
			LocationID:  asset.RoomID(locationID),
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			defer r.Body.Close()

			var createReq rest.PlayerCreateRequest
			err = json.Unmarshal(body, &createReq)
			assert.Nil(t, err)
			assert.Equal(t, createReq, rest.PlayerCreateRequest{PlayerRequest: rest.PlayerRequest{
				Name:        name,
				Description: desc,
				HomeID:      home,
				LocationID:  loc,
			}})

			err = json.NewEncoder(w).Encode(rest.PlayerResponse{Player: rPlayer})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		player, err := c.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
			},
		})

		assert.Nil(t, err)
		assert.Compare(t, player, aPlayer, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestUpdatePlayer(t *testing.T) {
	const (
		name = "name"
		desc = "desc"
		id   = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		home = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		loc  = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
	)

	var (
		ctx        = context.Background()
		playerID   = asset.PlayerID(uuid.MustParse(id))
		homeID     = asset.RoomID(uuid.MustParse(home))
		locationID = asset.RoomID(uuid.MustParse(loc))
	)

	t.Run("player change failure", func(t *testing.T) {
		c := client.Client{}

		_, err := c.UpdatePlayer(ctx, playerID, asset.PlayerUpdate{PlayerChange: asset.PlayerChange{
			Name: name,
		}})

		assert.Error(t, err, `failed to update player: attempted to send empty description in request`)
	})

	t.Run("update request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.UpdatePlayer(ctx, playerID, asset.PlayerUpdate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
			},
		})

		assert.Contains(t, err.Error(), `failed to update player: parse "1234:bad url/v1/player/db81f22a-90cf-48a7-93a2-94de93a9b48f": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.UpdatePlayer(ctx, playerID, asset.PlayerUpdate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
			},
		})

		assert.Error(t, err, `failed to update player: 500, Internal Server Error`)
	})

	t.Run("translate player failure", func(t *testing.T) {
		player := rest.Player{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.PlayerResponse{Player: player})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.UpdatePlayer(ctx, playerID, asset.PlayerUpdate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
			},
		})

		assert.Error(t, err, `failed to update player: received invalid player ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		var (
			created = arcade.Timestamp{Time: time.Now().UTC()}
			updated = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rPlayer := rest.Player{
			ID:          id,
			Name:        name,
			Description: desc,
			HomeID:      home,
			LocationID:  loc,
			Created:     created,
			Updated:     updated,
		}

		aPlayer := &asset.Player{
			ID:          playerID,
			Name:        name,
			Description: desc,
			HomeID:      homeID,
			LocationID:  locationID,
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			defer r.Body.Close()

			var createReq rest.PlayerCreateRequest
			err = json.Unmarshal(body, &createReq)
			assert.Nil(t, err)
			assert.Equal(t, createReq, rest.PlayerCreateRequest{PlayerRequest: rest.PlayerRequest{
				Name:        name,
				Description: desc,
				HomeID:      home,
				LocationID:  loc,
			}})

			err = json.NewEncoder(w).Encode(rest.PlayerResponse{Player: rPlayer})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		player, err := c.UpdatePlayer(ctx, playerID, asset.PlayerUpdate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: desc,
				HomeID:      homeID,
				LocationID:  locationID,
			},
		})

		assert.Nil(t, err)
		assert.Compare(t, player, aPlayer, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestRemovePlayer(t *testing.T) {
	var (
		ctx = context.Background()
		id  = asset.PlayerID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		err := c.RemovePlayer(ctx, id)

		assert.Contains(t, err.Error(), `failed to remove player: parse "1234:bad url/v1/player/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		err := c.RemovePlayer(ctx, id)

		assert.Error(t, err, `failed to remove player: 500, Internal Server Error`)
	})

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := client.New(server.URL)

		err := c.RemovePlayer(ctx, id)

		assert.Nil(t, err)
	})
}

func TestTranslatePlayer(t *testing.T) {
	const (
		badID   = "bad uuid"
		goodID  = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		badType = "bad type"
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			name    string
			rPlayer rest.Player
			err     string
		}{
			{
				name: "bad id",
				rPlayer: rest.Player{
					ID: badID,
				},
				err: "received invalid player ID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad homeID",
				rPlayer: rest.Player{
					ID:     goodID,
					HomeID: badID,
				},
				err: "received invalid player homeID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad locationID",
				rPlayer: rest.Player{
					ID:         goodID,
					HomeID:     goodID,
					LocationID: badID,
				},
				err: "received invalid player locationID: 'bad uuid': invalid UUID length: 8",
			},
		}

		for _, test := range tests {
			t.Logf("name: %s", test.name)
			i, err := client.TranslatePlayer(test.rPlayer)
			assert.Nil(t, i)
			assert.Error(t, err, test.err)
		}
	})

	t.Run("success", func(t *testing.T) {
		const (
			id   = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			name = "name"
			desc = "desc"
			home = "7f5908a2-3f99-4e21-a621-d369cff3b061"
			loc  = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
		)
		var (
			playerID   = asset.PlayerID(uuid.MustParse(id))
			homeID     = asset.RoomID(uuid.MustParse(home))
			locationID = asset.RoomID(uuid.MustParse(loc))
			created    = arcade.Timestamp{Time: time.Now().UTC()}
			updated    = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rPlayer := rest.Player{
			ID:          id,
			Name:        name,
			Description: desc,
			HomeID:      home,
			LocationID:  loc,
			Created:     created,
			Updated:     updated,
		}
		aPlayer := &asset.Player{
			ID:          playerID,
			Name:        name,
			Description: desc,
			HomeID:      homeID,
			LocationID:  locationID,
			Created:     created,
			Updated:     updated,
		}

		p, err := client.TranslatePlayer(rPlayer)
		assert.Nil(t, err)
		assert.Compare(t, p, aPlayer)
	})
}

func TestTranslatePlayerChange(t *testing.T) {
	const (
		name = "name"
		desc = "desc"
		home = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		loc  = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
	)

	var (
		homeID     = asset.RoomID(uuid.MustParse(home))
		locationID = asset.RoomID(uuid.MustParse(loc))
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			change asset.PlayerChange
			err    string
		}{
			{
				change: asset.PlayerChange{},
				err:    "attempted to send empty name in request",
			},
			{
				change: asset.PlayerChange{
					Name: name,
				},
				err: "attempted to send empty description in request",
			},
		}

		for _, test := range tests {
			_, err := client.TranslatePlayerChange(test.change)
			assert.Error(t, err, test.err)
		}
	})

	t.Run("success", func(t *testing.T) {
		change := asset.PlayerChange{
			Name:        name,
			Description: desc,
			HomeID:      homeID,
			LocationID:  locationID,
		}
		req, err := client.TranslatePlayerChange(change)
		assert.Nil(t, err)
		assert.Equal(t, req, rest.PlayerRequest{
			Name:        name,
			Description: desc,
			HomeID:      home,
			LocationID:  loc,
		})
	})
}
