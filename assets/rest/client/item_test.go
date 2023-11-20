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

	"arcadium.dev/core/assert"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/arcade/assets/rest"
	"arcadium.dev/arcade/assets/rest/client"
)

func TestListItems(t *testing.T) {
	ctx := context.Background()

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.ListItems(ctx, assets.ItemFilter{})

		assert.Contains(t, err.Error(), `failed to list items: parse "1234:bad url/v1/items": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListItems(ctx, assets.ItemFilter{})

		assert.Error(t, err, `failed to list items: 500, Internal Server Error`)
	})

	t.Run("response decode failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ foo`)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListItems(ctx, assets.ItemFilter{})

		assert.Error(t, err, `failed to list items: invalid character 'f' looking for beginning of object key string`)
	})

	t.Run("translate item failure", func(t *testing.T) {
		items := []rest.Item{
			{ID: "bad uuid"},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.ItemsResponse{Items: items})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListItems(ctx, assets.ItemFilter{})

		assert.Error(t, err, `failed to list items: received invalid item ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id   = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			name = "name"
			desc = "desc"
		)
		var (
			u       = uuid.MustParse(id)
			created = assets.Timestamp{Time: time.Now().UTC()}
			updated = assets.Timestamp{Time: time.Now().UTC()}
		)

		rItems := []rest.Item{
			{
				ID:          id,
				Name:        name,
				Description: desc,
				OwnerID:     id,
				LocationID: rest.ItemLocationID{
					ID:   id,
					Type: "room",
				},
				Created: created,
				Updated: updated,
			},
		}

		aItems := []*assets.Item{
			{
				ID:          assets.ItemID(u),
				Name:        name,
				Description: desc,
				OwnerID:     assets.PlayerID(u),
				LocationID:  assets.RoomID(u),
				Created:     created,
				Updated:     updated,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.ItemsResponse{Items: rItems})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		filter := assets.ItemFilter{
			OwnerID:    assets.PlayerID(u),
			LocationID: assets.RoomID(u),
			Offset:     10,
			Limit:      10,
		}

		items, err := c.ListItems(ctx, filter)

		assert.Nil(t, err)
		assert.Equal(t, len(items), 1)
		assert.Compare(t, *items[0], *aItems[0], cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestGetItem(t *testing.T) {
	ctx := context.Background()

	var (
		id = assets.ItemID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.GetItem(ctx, id)

		assert.Contains(t, err.Error(), `failed to get item: parse "1234:bad url/v1/items/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetItem(ctx, id)

		assert.Error(t, err, `failed to get item: 500, Internal Server Error`)
	})

	t.Run("response decode failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ foo`)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetItem(ctx, id)

		assert.Error(t, err, `failed to get item: invalid character 'f' looking for beginning of object key string`)
	})

	t.Run("translate item failure", func(t *testing.T) {
		item := rest.Item{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.ItemResponse{Item: item})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetItem(ctx, id)

		assert.Error(t, err, `failed to get item: received invalid item ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id   = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			name = "name"
			desc = "desc"
		)
		var (
			u       = uuid.MustParse(id)
			created = assets.Timestamp{Time: time.Now().UTC()}
			updated = assets.Timestamp{Time: time.Now().UTC()}
		)

		rItem := rest.Item{
			ID:          id,
			Name:        name,
			Description: desc,
			OwnerID:     id,
			LocationID: rest.ItemLocationID{
				ID:   id,
				Type: "room",
			},
			Created: created,
			Updated: updated,
		}

		aItem := &assets.Item{
			ID:          assets.ItemID(u),
			Name:        name,
			Description: desc,
			OwnerID:     assets.PlayerID(u),
			LocationID:  assets.RoomID(u),
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.ItemResponse{Item: rItem})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		item, err := c.GetItem(ctx, assets.ItemID(u))

		assert.Nil(t, err)
		assert.Compare(t, item, aItem, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestCreateItem(t *testing.T) {
	const (
		name  = "name"
		desc  = "desc"
		owner = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		loc   = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
	)

	var (
		ctx        = context.Background()
		ownerID    = assets.PlayerID(uuid.MustParse(owner))
		locationID = assets.RoomID(uuid.MustParse(loc))
	)

	t.Run("item change failure", func(t *testing.T) {
		c := client.Client{}

		_, err := c.CreateItem(ctx, assets.ItemCreate{ItemChange: assets.ItemChange{Name: ""}})

		assert.Error(t, err, `failed to create item: attempted to send empty name in request`)
	})

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.CreateItem(ctx, assets.ItemCreate{
			ItemChange: assets.ItemChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
			},
		})

		assert.Contains(t, err.Error(), `failed to create item: parse "1234:bad url/v1/items": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.CreateItem(ctx, assets.ItemCreate{
			ItemChange: assets.ItemChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
			},
		})

		assert.Error(t, err, `failed to create item: 500, Internal Server Error`)
	})

	t.Run("translate item failure", func(t *testing.T) {
		item := rest.Item{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.ItemResponse{Item: item})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.CreateItem(ctx, assets.ItemCreate{
			ItemChange: assets.ItemChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
			},
		})

		assert.Error(t, err, `failed to create item: received invalid item ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		)
		var (
			itemID  = assets.ItemID(uuid.MustParse(id))
			created = assets.Timestamp{Time: time.Now().UTC()}
			updated = assets.Timestamp{Time: time.Now().UTC()}
		)

		rItem := rest.Item{
			ID:          id,
			Name:        name,
			Description: desc,
			OwnerID:     owner,
			LocationID: rest.ItemLocationID{
				ID:   loc,
				Type: "room",
			},
			Created: created,
			Updated: updated,
		}

		aItem := &assets.Item{
			ID:          itemID,
			Name:        name,
			Description: desc,
			OwnerID:     assets.PlayerID(ownerID),
			LocationID:  assets.RoomID(locationID),
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			defer r.Body.Close()

			var createReq rest.ItemCreateRequest
			err = json.Unmarshal(body, &createReq)
			assert.Nil(t, err)
			assert.Equal(t, createReq, rest.ItemCreateRequest{ItemRequest: rest.ItemRequest{
				Name:        name,
				Description: desc,
				OwnerID:     owner,
				LocationID: rest.ItemLocationID{
					ID:   loc,
					Type: "room",
				},
			}})

			err = json.NewEncoder(w).Encode(rest.ItemResponse{Item: rItem})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		item, err := c.CreateItem(ctx, assets.ItemCreate{
			ItemChange: assets.ItemChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
			},
		})

		assert.Nil(t, err)
		assert.Compare(t, item, aItem, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestUpdateItem(t *testing.T) {
	const (
		name  = "name"
		desc  = "desc"
		owner = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		loc   = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
	)

	var (
		ctx        = context.Background()
		ownerID    = assets.PlayerID(uuid.MustParse(owner))
		locationID = assets.RoomID(uuid.MustParse(loc))
	)

	var (
		id = assets.ItemID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("item change failure", func(t *testing.T) {
		c := client.Client{}

		_, err := c.UpdateItem(ctx, id, assets.ItemUpdate{ItemChange: assets.ItemChange{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID,
		}})

		assert.Error(t, err, `failed to update item: attempted to send empty locationID in request`)
	})

	t.Run("update request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.UpdateItem(ctx, id, assets.ItemUpdate{
			ItemChange: assets.ItemChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
			},
		})

		assert.Contains(t, err.Error(), `failed to update item: parse "1234:bad url/v1/items/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.UpdateItem(ctx, id, assets.ItemUpdate{
			ItemChange: assets.ItemChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
			},
		})

		assert.Error(t, err, `failed to update item: 500, Internal Server Error`)
	})

	t.Run("translate item failure", func(t *testing.T) {
		item := rest.Item{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.ItemResponse{Item: item})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.UpdateItem(ctx, id, assets.ItemUpdate{
			ItemChange: assets.ItemChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
			},
		})

		assert.Error(t, err, `failed to update item: received invalid item ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		)
		var (
			itemID  = assets.ItemID(uuid.MustParse(id))
			created = assets.Timestamp{Time: time.Now().UTC()}
			updated = assets.Timestamp{Time: time.Now().UTC()}
		)

		rItem := rest.Item{
			ID:          id,
			Name:        name,
			Description: desc,
			OwnerID:     owner,
			LocationID: rest.ItemLocationID{
				ID:   loc,
				Type: "room",
			},
			Created: created,
			Updated: updated,
		}

		aItem := &assets.Item{
			ID:          itemID,
			Name:        name,
			Description: desc,
			OwnerID:     ownerID,
			LocationID:  locationID,
			Created:     created,
			Updated:     updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			defer r.Body.Close()

			var createReq rest.ItemCreateRequest
			err = json.Unmarshal(body, &createReq)
			assert.Nil(t, err)
			assert.Equal(t, createReq, rest.ItemCreateRequest{ItemRequest: rest.ItemRequest{
				Name:        name,
				Description: desc,
				OwnerID:     owner,
				LocationID: rest.ItemLocationID{
					ID:   loc,
					Type: "room",
				},
			}})

			err = json.NewEncoder(w).Encode(rest.ItemResponse{Item: rItem})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		item, err := c.UpdateItem(ctx, itemID, assets.ItemUpdate{
			ItemChange: assets.ItemChange{
				Name:        name,
				Description: desc,
				OwnerID:     ownerID,
				LocationID:  locationID,
			},
		})

		assert.Nil(t, err)
		assert.Compare(t, item, aItem, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestRemoveItem(t *testing.T) {
	ctx := context.Background()

	var (
		id = assets.ItemID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		err := c.RemoveItem(ctx, id)

		assert.Contains(t, err.Error(), `failed to remove item: parse "1234:bad url/v1/items/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		err := c.RemoveItem(ctx, id)

		assert.Error(t, err, `failed to remove item: 500, Internal Server Error`)
	})

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := client.New(server.URL)

		err := c.RemoveItem(ctx, id)

		assert.Nil(t, err)
	})
}

func TestTranslateItem(t *testing.T) {
	const (
		badID   = "bad uuid"
		goodID  = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		badType = "bad type"
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			name  string
			rItem rest.Item
			err   string
		}{
			{
				name: "bad id",
				rItem: rest.Item{
					ID: badID,
				},
				err: "received invalid item ID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad ownerID",
				rItem: rest.Item{
					ID:      goodID,
					OwnerID: badID,
				},
				err: "received invalid item ownerID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad locationID.ID",
				rItem: rest.Item{
					ID:      goodID,
					OwnerID: goodID,
					LocationID: rest.ItemLocationID{
						ID: badID,
					},
				},
				err: "received invalid item locationID.ID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad locationID.Type",
				rItem: rest.Item{
					ID:      goodID,
					OwnerID: goodID,
					LocationID: rest.ItemLocationID{
						ID:   goodID,
						Type: badType,
					},
				},
				err: "received invalid item locationID.Type: 'bad type'",
			},
		}

		for _, test := range tests {
			t.Logf("name: %s", test.name)
			i, err := client.TranslateItem(test.rItem)
			assert.Nil(t, i)
			assert.Error(t, err, test.err)
		}
	})

	t.Run("success", func(t *testing.T) {
		const (
			ids  = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			name = "name"
			desc = "desc"
		)
		var (
			id      = uuid.MustParse(ids)
			created = assets.Timestamp{Time: time.Now().UTC()}
			updated = assets.Timestamp{Time: time.Now().UTC()}
		)

		tests := []struct {
			locationType string
			locationID   assets.ItemLocationID
		}{
			{
				locationType: "room",
				locationID:   assets.RoomID(id),
			},
			{
				locationType: "player",
				locationID:   assets.PlayerID(id),
			},
			{
				locationType: "item",
				locationID:   assets.ItemID(id),
			},
		}

		for _, test := range tests {
			t.Logf("location: %s", test.locationType)
			rItem := rest.Item{
				ID:          ids,
				Name:        name,
				Description: desc,
				OwnerID:     ids,
				LocationID: rest.ItemLocationID{
					ID:   ids,
					Type: test.locationType,
				},
				Created: created,
				Updated: updated,
			}
			aItem := &assets.Item{
				ID:          assets.ItemID(id),
				Name:        name,
				Description: desc,
				OwnerID:     assets.PlayerID(id),
				LocationID:  test.locationID,
				Created:     created,
				Updated:     updated,
			}
			i, err := client.TranslateItem(rItem)
			assert.Nil(t, err)
			assert.Compare(t, i, aItem)
		}
	})
}

func TestTranslateItemChange(t *testing.T) {
	const (
		name  = "name"
		desc  = "desc"
		owner = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		loc   = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
	)

	var (
		ownerID    = assets.PlayerID(uuid.MustParse(owner))
		locationID = assets.RoomID(uuid.MustParse(loc))
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			change assets.ItemChange
			err    string
		}{
			{
				change: assets.ItemChange{},
				err:    "attempted to send empty name in request",
			},
			{
				change: assets.ItemChange{
					Name: name,
				},
				err: "attempted to send empty description in request",
			},
			{
				change: assets.ItemChange{
					Name:        name,
					Description: desc,
				},
				err: "attempted to send empty locationID in request",
			},
		}

		for _, test := range tests {
			_, err := client.TranslateItemChange(test.change)
			assert.Error(t, err, test.err)
		}
	})

	t.Run("success", func(t *testing.T) {
		change := assets.ItemChange{
			Name:        name,
			Description: desc,
			OwnerID:     ownerID,
			LocationID:  locationID,
		}
		req, err := client.TranslateItemChange(change)
		assert.Nil(t, err)
		assert.Equal(t, req, rest.ItemRequest{
			Name:        name,
			Description: desc,
			OwnerID:     owner,
			LocationID: rest.ItemLocationID{
				ID:   loc,
				Type: "room",
			},
		})
	})
}
