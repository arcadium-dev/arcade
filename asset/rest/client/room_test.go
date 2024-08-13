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

	"arcadium.dev/core/assert"
	"arcadium.dev/core/require"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
	"arcadium.dev/arcade/asset/rest/client"
)

func TestListRooms(t *testing.T) {
	ctx := context.Background()

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.ListRooms(ctx, asset.RoomFilter{})

		assert.Contains(t, err.Error(), `failed to list rooms: parse "1234:bad url/v1/room": first path segment in URL cannot contain colon`)
	})

	t.Run("max limit failure", func(t *testing.T) {
		c := client.New("https://example.com")

		_, err := c.ListRooms(ctx, asset.RoomFilter{Limit: 1000})

		assert.Contains(t, err.Error(), `failed to list rooms: room filter limit 1000 exceeds maximum 100`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListRooms(ctx, asset.RoomFilter{})

		assert.Error(t, err, `failed to list rooms: 500, Internal Server Error`)
	})

	t.Run("response decode failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ foo`)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListRooms(ctx, asset.RoomFilter{})

		assert.Error(t, err, `failed to list rooms: invalid character 'f' looking for beginning of object key string`)
	})

	t.Run("translate room failure", func(t *testing.T) {
		rooms := []rest.Room{
			{ID: "bad uuid"},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.RoomsResponse{Rooms: rooms})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListRooms(ctx, asset.RoomFilter{})

		assert.Error(t, err, `failed to list rooms: received invalid room ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id     = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			owner  = "1f290a67-ef2c-455b-aad7-6a3e72276ab5"
			parent = "8f314204-49f5-44c0-83f1-1e3a24eec3ad"
			name   = "name"
			desc   = "desc"
		)
		var (
			roomID   = asset.RoomID(uuid.MustParse(id))
			ownerID  = asset.PlayerID(uuid.MustParse(owner))
			parentID = asset.RoomID(uuid.MustParse(parent))
			created  = arcade.Timestamp{Time: time.Now().UTC()}
			updated  = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rRooms := []rest.Room{
			{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     owner,
				ParentID:    parent,
				Created:     created,
				Updated:     updated,
			},
		}

		aRooms := []*asset.Room{
			{
				ID:          roomID,
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
				Created:     created,
				Updated:     updated,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			require.Equal(t, len(q["ownerID"]), 1)
			assert.Equal(t, q["ownerID"][0], owner)
			require.Equal(t, len(q["parentID"]), 1)
			assert.Equal(t, q["parentID"][0], parent)
			require.Equal(t, len(q["offset"]), 1)
			assert.Equal(t, q["offset"][0], "10")
			require.Equal(t, len(q["limit"]), 1)
			assert.Equal(t, q["limit"][0], "10")

			err := json.NewEncoder(w).Encode(rest.RoomsResponse{Rooms: rRooms})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		filter := asset.RoomFilter{
			OwnerID:  ownerID,
			ParentID: parentID,
			Offset:   10,
			Limit:    10,
		}

		rooms, err := c.ListRooms(ctx, filter)

		assert.Nil(t, err)
		assert.Equal(t, len(rooms), 1)
		assert.Compare(t, *rooms[0], *aRooms[0], cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestGetRoom(t *testing.T) {
	var (
		ctx = context.Background()
		id  = asset.RoomID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.GetRoom(ctx, id)

		assert.Contains(t, err.Error(), `failed to get room: parse "1234:bad url/v1/room/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetRoom(ctx, id)

		assert.Error(t, err, `failed to get room: 500, Internal Server Error`)
	})

	t.Run("response decode failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ foo`)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetRoom(ctx, id)

		assert.Error(t, err, `failed to get room: invalid character 'f' looking for beginning of object key string`)
	})

	t.Run("translate room failure", func(t *testing.T) {
		room := rest.Room{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.RoomResponse{Room: room})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetRoom(ctx, id)

		assert.Error(t, err, `failed to get room: received invalid room ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id     = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			owner  = "1f290a67-ef2c-455b-aad7-6a3e72276ab5"
			parent = "8f314204-49f5-44c0-83f1-1e3a24eec3ad"
			name   = "name"
			desc   = "desc"
		)
		var (
			roomID   = asset.RoomID(uuid.MustParse(id))
			ownerID  = asset.PlayerID(uuid.MustParse(owner))
			parentID = asset.RoomID(uuid.MustParse(parent))
			created  = arcade.Timestamp{Time: time.Now().UTC()}
			updated  = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rRoom := rest.Room{
			ID:          id,
			Name:        name,
			Description: desc,
			OwnerID:     owner,
			ParentID:    parent,
			Created:     created,
			Updated:     updated,
		}

		aRoom := &asset.Room{
			ID:          roomID,
			Name:        name,
			Description: desc,
			OwnerID:     ownerID,
			ParentID:    parentID,
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.RoomResponse{Room: rRoom})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		room, err := c.GetRoom(ctx, roomID)

		assert.Nil(t, err)
		assert.Compare(t, room, aRoom, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestCreateRoom(t *testing.T) {
	const (
		name   = "name"
		desc   = "desc"
		id     = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		owner  = "1f290a67-ef2c-455b-aad7-6a3e72276ab5"
		parent = "8f314204-49f5-44c0-83f1-1e3a24eec3ad"
	)

	var (
		ctx      = context.Background()
		roomID   = asset.RoomID(uuid.MustParse(id))
		ownerID  = asset.PlayerID(uuid.MustParse(owner))
		parentID = asset.RoomID(uuid.MustParse(parent))
	)

	t.Run("room change failure", func(t *testing.T) {
		c := client.Client{}

		_, err := c.CreateRoom(ctx, asset.RoomCreate{RoomChange: asset.RoomChange{Name: ""}})

		assert.Error(t, err, `failed to create room: attempted to send empty name in request`)
	})

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
		})

		assert.Contains(t, err.Error(), `failed to create room: parse "1234:bad url/v1/room": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
		})

		assert.Error(t, err, `failed to create room: 500, Internal Server Error`)
	})

	t.Run("translate room failure", func(t *testing.T) {
		room := rest.Room{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.RoomResponse{Room: room})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
		})

		assert.Error(t, err, `failed to create room: received invalid room ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		var (
			created = arcade.Timestamp{Time: time.Now().UTC()}
			updated = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rRoom := rest.Room{
			ID:          id,
			Name:        name,
			Description: desc,
			OwnerID:     owner,
			ParentID:    parent,
			Created:     created,
			Updated:     updated,
		}

		aRoom := &asset.Room{
			ID:          roomID,
			Name:        name,
			Description: desc,
			OwnerID:     ownerID,
			ParentID:    parentID,
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			defer r.Body.Close()

			var createReq rest.RoomCreateRequest
			err = json.Unmarshal(body, &createReq)
			assert.Nil(t, err)
			assert.Equal(t, createReq, rest.RoomCreateRequest{RoomRequest: rest.RoomRequest{
				Name:        name,
				Description: desc,
				OwnerID:     owner,
				ParentID:    parent,
			}})

			err = json.NewEncoder(w).Encode(rest.RoomResponse{Room: rRoom})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		room, err := c.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
		})

		assert.Nil(t, err)
		assert.Compare(t, room, aRoom, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestUpdateRoom(t *testing.T) {
	const (
		name   = "name"
		desc   = "desc"
		id     = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		owner  = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		parent = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
	)

	var (
		ctx      = context.Background()
		roomID   = asset.RoomID(uuid.MustParse(id))
		ownerID  = asset.PlayerID(uuid.MustParse(owner))
		parentID = asset.RoomID(uuid.MustParse(parent))
	)

	t.Run("room change failure", func(t *testing.T) {
		c := client.Client{}

		_, err := c.UpdateRoom(ctx, roomID, asset.RoomUpdate{RoomChange: asset.RoomChange{
			Name: name,
		}})

		assert.Error(t, err, `failed to update room: attempted to send empty description in request`)
	})

	t.Run("update request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.UpdateRoom(ctx, roomID, asset.RoomUpdate{
			RoomChange: asset.RoomChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
		})

		assert.Contains(t, err.Error(), `failed to update room: parse "1234:bad url/v1/room/db81f22a-90cf-48a7-93a2-94de93a9b48f": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.UpdateRoom(ctx, roomID, asset.RoomUpdate{
			RoomChange: asset.RoomChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
		})

		assert.Error(t, err, `failed to update room: 500, Internal Server Error`)
	})

	t.Run("translate room failure", func(t *testing.T) {
		room := rest.Room{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.RoomResponse{Room: room})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.UpdateRoom(ctx, roomID, asset.RoomUpdate{
			RoomChange: asset.RoomChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
		})

		assert.Error(t, err, `failed to update room: received invalid room ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		var (
			created = arcade.Timestamp{Time: time.Now().UTC()}
			updated = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rRoom := rest.Room{
			ID:          id,
			Name:        name,
			Description: desc,
			OwnerID:     owner,
			ParentID:    parent,
			Created:     created,
			Updated:     updated,
		}

		aRoom := &asset.Room{
			ID:          roomID,
			Name:        name,
			Description: desc,
			OwnerID:     ownerID,
			ParentID:    parentID,
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			defer r.Body.Close()

			var createReq rest.RoomCreateRequest
			err = json.Unmarshal(body, &createReq)
			assert.Nil(t, err)
			assert.Equal(t, createReq, rest.RoomCreateRequest{RoomRequest: rest.RoomRequest{
				Name:        name,
				Description: desc,
				OwnerID:     owner,
				ParentID:    parent,
			}})

			err = json.NewEncoder(w).Encode(rest.RoomResponse{Room: rRoom})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		room, err := c.UpdateRoom(ctx, roomID, asset.RoomUpdate{
			RoomChange: asset.RoomChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				ParentID:    parentID,
			},
		})

		assert.Nil(t, err)
		assert.Compare(t, room, aRoom, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestRemoveRoom(t *testing.T) {
	var (
		ctx = context.Background()
		id  = asset.RoomID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		err := c.RemoveRoom(ctx, id)

		assert.Contains(t, err.Error(), `failed to remove room: parse "1234:bad url/v1/room/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		err := c.RemoveRoom(ctx, id)

		assert.Error(t, err, `failed to remove room: 500, Internal Server Error`)
	})

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := client.New(server.URL)

		err := c.RemoveRoom(ctx, id)

		assert.Nil(t, err)
	})
}

func TestTranslateRoom(t *testing.T) {
	const (
		badID   = "bad uuid"
		goodID  = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		badType = "bad type"
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			name  string
			rRoom rest.Room
			err   string
		}{
			{
				name: "bad id",
				rRoom: rest.Room{
					ID: badID,
				},
				err: "received invalid room ID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad ownerID",
				rRoom: rest.Room{
					ID:      goodID,
					OwnerID: badID,
				},
				err: "received invalid room ownerID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad parentID",
				rRoom: rest.Room{
					ID:       goodID,
					OwnerID:  goodID,
					ParentID: badID,
				},
				err: "received invalid room parentID: 'bad uuid': invalid UUID length: 8",
			},
		}

		for _, test := range tests {
			t.Logf("name: %s", test.name)
			i, err := client.TranslateRoom(test.rRoom)
			assert.Nil(t, i)
			assert.Error(t, err, test.err)
		}
	})

	t.Run("success", func(t *testing.T) {
		const (
			name   = "name"
			desc   = "desc"
			id     = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			owner  = "7f5908a2-3f99-4e21-a621-d369cff3b061"
			parent = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
		)
		var (
			roomID   = asset.RoomID(uuid.MustParse(id))
			ownerID  = asset.PlayerID(uuid.MustParse(owner))
			parentID = asset.RoomID(uuid.MustParse(parent))
			created  = arcade.Timestamp{Time: time.Now().UTC()}
			updated  = arcade.Timestamp{Time: time.Now().UTC()}
		)

		rRoom := rest.Room{
			ID:          id,
			Name:        name,
			Description: desc,
			OwnerID:     owner,
			ParentID:    parent,
			Created:     created,
			Updated:     updated,
		}
		aRoom := &asset.Room{
			ID:          roomID,
			Name:        name,
			Description: desc,
			OwnerID:     ownerID,
			ParentID:    parentID,
			Created:     created,
			Updated:     updated,
		}

		p, err := client.TranslateRoom(rRoom)
		assert.Nil(t, err)
		assert.Compare(t, p, aRoom)
	})
}

func TestTranslateRoomChange(t *testing.T) {
	const (
		name   = "name"
		desc   = "desc"
		owner  = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		parent = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
	)

	var (
		ownerID  = asset.PlayerID(uuid.MustParse(owner))
		parentID = asset.RoomID(uuid.MustParse(parent))
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			change asset.RoomChange
			err    string
		}{
			{
				change: asset.RoomChange{},
				err:    "attempted to send empty name in request",
			},
			{
				change: asset.RoomChange{
					Name: name,
				},
				err: "attempted to send empty description in request",
			},
		}

		for _, test := range tests {
			_, err := client.TranslateRoomChange(test.change)
			assert.Error(t, err, test.err)
		}
	})

	t.Run("success", func(t *testing.T) {
		change := asset.RoomChange{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID,
			ParentID:    parentID,
		}
		req, err := client.TranslateRoomChange(change)
		assert.Nil(t, err)
		assert.Equal(t, req, rest.RoomRequest{
			Name:        name,
			Description: desc,
			OwnerID:     owner,
			ParentID:    parent,
		})
	})
}
