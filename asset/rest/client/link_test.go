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

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
	"arcadium.dev/arcade/asset/rest/client"
	"arcadium.dev/core/assert"
	"arcadium.dev/core/require"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

func TestListLinks(t *testing.T) {
	ctx := context.Background()

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.ListLinks(ctx, asset.LinkFilter{})

		assert.Contains(t, err.Error(), `failed to list links: parse "1234:bad url/v1/link": first path segment in URL cannot contain colon`)
	})

	t.Run("max limit failure", func(t *testing.T) {
		c := client.New("https://example.com")

		_, err := c.ListLinks(ctx, asset.LinkFilter{Limit: 1000})

		assert.Contains(t, err.Error(), `failed to list links: link filter limit 1000 exceeds maximum 100`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListLinks(ctx, asset.LinkFilter{})

		assert.Error(t, err, `failed to list links: 500, Internal Server Error`)
	})

	t.Run("response decode failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ foo`)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListLinks(ctx, asset.LinkFilter{})

		assert.Error(t, err, `failed to list links: invalid character 'f' looking for beginning of object key string`)
	})

	t.Run("translate link failure", func(t *testing.T) {
		links := []rest.Link{
			{ID: "bad uuid"},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.LinksResponse{Links: links})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.ListLinks(ctx, asset.LinkFilter{})

		assert.Error(t, err, `failed to list links: received invalid link ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id          = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			owner       = "1f290a67-ef2c-455b-aad7-6a3e72276ab5"
			location    = "8f314204-49f5-44c0-83f1-1e3a24eec3ad"
			destination = "5ce8aab9-4c49-4870-98f0-4bba0316727a"
			name        = "name"
			desc        = "desc"
		)
		var (
			linkID        = asset.LinkID(uuid.MustParse(id))
			ownerID       = asset.PlayerID(uuid.MustParse(owner))
			locationID    = asset.RoomID(uuid.MustParse(location))
			destinationID = asset.RoomID(uuid.MustParse(destination))
			created       = asset.Timestamp{Time: time.Now().UTC()}
			updated       = asset.Timestamp{Time: time.Now().UTC()}
		)

		rLinks := []rest.Link{
			{
				ID:            id,
				Name:          name,
				Description:   desc,
				OwnerID:       owner,
				LocationID:    location,
				DestinationID: destination,
				Created:       created,
				Updated:       updated,
			},
		}

		aLinks := []*asset.Link{
			{
				ID:            linkID,
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
				Created:       created,
				Updated:       updated,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			require.Equal(t, len(q["ownerID"]), 1)
			assert.Equal(t, q["ownerID"][0], owner)
			require.Equal(t, len(q["locationID"]), 1)
			assert.Equal(t, q["locationID"][0], location)
			require.Equal(t, len(q["destinationID"]), 1)
			assert.Equal(t, q["destinationID"][0], destination)
			require.Equal(t, len(q["offset"]), 1)
			assert.Equal(t, q["offset"][0], "10")
			require.Equal(t, len(q["limit"]), 1)
			assert.Equal(t, q["limit"][0], "10")

			err := json.NewEncoder(w).Encode(rest.LinksResponse{Links: rLinks})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		filter := asset.LinkFilter{
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
			Offset:        10,
			Limit:         10,
		}

		links, err := c.ListLinks(ctx, filter)

		assert.Nil(t, err)
		assert.Equal(t, len(links), 1)
		assert.Compare(t, *links[0], *aLinks[0], cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestGetLink(t *testing.T) {
	var (
		ctx = context.Background()
		id  = asset.LinkID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.GetLink(ctx, id)

		assert.Contains(t, err.Error(), `failed to get link: parse "1234:bad url/v1/link/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetLink(ctx, id)

		assert.Error(t, err, `failed to get link: 500, Internal Server Error`)
	})

	t.Run("response decode failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ foo`)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetLink(ctx, id)

		assert.Error(t, err, `failed to get link: invalid character 'f' looking for beginning of object key string`)
	})

	t.Run("translate link failure", func(t *testing.T) {
		link := rest.Link{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.LinkResponse{Link: link})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.GetLink(ctx, id)

		assert.Error(t, err, `failed to get link: received invalid link ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		const (
			id          = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			owner       = "1f290a67-ef2c-455b-aad7-6a3e72276ab5"
			location    = "8f314204-49f5-44c0-83f1-1e3a24eec3ad"
			destination = "5ce8aab9-4c49-4870-98f0-4bba0316727a"
			name        = "name"
			desc        = "desc"
		)
		var (
			linkID        = asset.LinkID(uuid.MustParse(id))
			ownerID       = asset.PlayerID(uuid.MustParse(owner))
			locationID    = asset.RoomID(uuid.MustParse(location))
			destinationID = asset.RoomID(uuid.MustParse(destination))
			created       = asset.Timestamp{Time: time.Now().UTC()}
			updated       = asset.Timestamp{Time: time.Now().UTC()}
		)

		rLink := rest.Link{
			ID:            id,
			Name:          name,
			Description:   desc,
			OwnerID:       owner,
			LocationID:    location,
			DestinationID: destination,
			Created:       created,
			Updated:       updated,
		}

		aLink := &asset.Link{
			ID:            linkID,
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
			Created:       created,
			Updated:       updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.LinkResponse{Link: rLink})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		link, err := c.GetLink(ctx, linkID)

		assert.Nil(t, err)
		assert.Compare(t, link, aLink, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestCreateLink(t *testing.T) {
	const (
		name        = "name"
		desc        = "desc"
		id          = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		owner       = "1f290a67-ef2c-455b-aad7-6a3e72276ab5"
		location    = "8f314204-49f5-44c0-83f1-1e3a24eec3ad"
		destination = "5ce8aab9-4c49-4870-98f0-4bba0316727a"
	)

	var (
		ctx           = context.Background()
		linkID        = asset.LinkID(uuid.MustParse(id))
		ownerID       = asset.PlayerID(uuid.MustParse(owner))
		locationID    = asset.RoomID(uuid.MustParse(location))
		destinationID = asset.RoomID(uuid.MustParse(destination))
	)

	t.Run("link change failure", func(t *testing.T) {
		c := client.Client{}

		_, err := c.CreateLink(ctx, asset.LinkCreate{LinkChange: asset.LinkChange{Name: ""}})

		assert.Error(t, err, `failed to create link: attempted to send empty name in request`)
	})

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.CreateLink(ctx, asset.LinkCreate{
			LinkChange: asset.LinkChange{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
			},
		})

		assert.Contains(t, err.Error(), `failed to create link: parse "1234:bad url/v1/link": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.CreateLink(ctx, asset.LinkCreate{
			LinkChange: asset.LinkChange{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
			},
		})

		assert.Error(t, err, `failed to create link: 500, Internal Server Error`)
	})

	t.Run("translate link failure", func(t *testing.T) {
		link := rest.Link{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.LinkResponse{Link: link})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.CreateLink(ctx, asset.LinkCreate{
			LinkChange: asset.LinkChange{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
			},
		})

		assert.Error(t, err, `failed to create link: received invalid link ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		var (
			created = asset.Timestamp{Time: time.Now().UTC()}
			updated = asset.Timestamp{Time: time.Now().UTC()}
		)

		rLink := rest.Link{
			ID:            id,
			Name:          name,
			Description:   desc,
			OwnerID:       owner,
			LocationID:    location,
			DestinationID: destination,
			Created:       created,
			Updated:       updated,
		}

		aLink := &asset.Link{
			ID:            linkID,
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
			Created:       created,
			Updated:       updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			defer r.Body.Close()

			var createReq rest.LinkCreateRequest
			err = json.Unmarshal(body, &createReq)
			assert.Nil(t, err)
			assert.Equal(t, createReq, rest.LinkCreateRequest{LinkRequest: rest.LinkRequest{
				Name:          name,
				Description:   desc,
				OwnerID:       owner,
				LocationID:    location,
				DestinationID: destination,
			}})

			err = json.NewEncoder(w).Encode(rest.LinkResponse{Link: rLink})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		link, err := c.CreateLink(ctx, asset.LinkCreate{
			LinkChange: asset.LinkChange{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
			},
		})

		assert.Nil(t, err)
		assert.Compare(t, link, aLink, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestUpdateLink(t *testing.T) {
	const (
		name        = "name"
		desc        = "desc"
		id          = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		owner       = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		location    = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
		destination = "5ce8aab9-4c49-4870-98f0-4bba0316727a"
	)

	var (
		ctx           = context.Background()
		linkID        = asset.LinkID(uuid.MustParse(id))
		ownerID       = asset.PlayerID(uuid.MustParse(owner))
		locationID    = asset.RoomID(uuid.MustParse(location))
		destinationID = asset.RoomID(uuid.MustParse(destination))
	)

	t.Run("link change failure", func(t *testing.T) {
		c := client.Client{}

		_, err := c.UpdateLink(ctx, linkID, asset.LinkUpdate{LinkChange: asset.LinkChange{
			Name: name,
		}})

		assert.Error(t, err, `failed to update link: attempted to send empty description in request`)
	})

	t.Run("update request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		_, err := c.UpdateLink(ctx, linkID, asset.LinkUpdate{
			LinkChange: asset.LinkChange{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
			},
		})

		assert.Contains(t, err.Error(), `failed to update link: parse "1234:bad url/v1/link/db81f22a-90cf-48a7-93a2-94de93a9b48f": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.UpdateLink(ctx, linkID, asset.LinkUpdate{
			LinkChange: asset.LinkChange{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
			},
		})

		assert.Error(t, err, `failed to update link: 500, Internal Server Error`)
	})

	t.Run("translate link failure", func(t *testing.T) {
		link := rest.Link{
			ID: "bad uuid",
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewEncoder(w).Encode(rest.LinkResponse{Link: link})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		_, err := c.UpdateLink(ctx, linkID, asset.LinkUpdate{
			LinkChange: asset.LinkChange{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
			},
		})

		assert.Error(t, err, `failed to update link: received invalid link ID: 'bad uuid': invalid UUID length: 8`)
	})

	t.Run("success", func(t *testing.T) {
		var (
			created = asset.Timestamp{Time: time.Now().UTC()}
			updated = asset.Timestamp{Time: time.Now().UTC()}
		)

		rLink := rest.Link{
			ID:            id,
			Name:          name,
			Description:   desc,
			OwnerID:       owner,
			LocationID:    location,
			DestinationID: destination,
			Created:       created,
			Updated:       updated,
		}

		aLink := &asset.Link{
			ID:            linkID,
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
			Created:       created,
			Updated:       updated,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			defer r.Body.Close()

			var createReq rest.LinkCreateRequest
			err = json.Unmarshal(body, &createReq)
			assert.Nil(t, err)
			assert.Equal(t, createReq, rest.LinkCreateRequest{LinkRequest: rest.LinkRequest{
				Name:          name,
				Description:   desc,
				OwnerID:       owner,
				LocationID:    location,
				DestinationID: destination,
			}})

			err = json.NewEncoder(w).Encode(rest.LinkResponse{Link: rLink})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		link, err := c.UpdateLink(ctx, linkID, asset.LinkUpdate{
			LinkChange: asset.LinkChange{
				Name:          name,
				Description:   desc,
				OwnerID:       ownerID,
				LocationID:    locationID,
				DestinationID: destinationID,
			},
		})

		assert.Nil(t, err)
		assert.Compare(t, link, aLink, cmpopts.EquateApproxTime(time.Duration(1*time.Microsecond)))
	})
}

func TestRemoveLink(t *testing.T) {
	var (
		ctx = context.Background()
		id  = asset.LinkID(uuid.MustParse("4efee5c1-01ac-41c6-a479-0ae59617482b"))
	)

	t.Run("create request failure", func(t *testing.T) {
		c := client.New("1234:bad url")

		err := c.RemoveLink(ctx, id)

		assert.Contains(t, err.Error(), `failed to remove link: parse "1234:bad url/v1/link/4efee5c1-01ac-41c6-a479-0ae59617482b": first path segment in URL cannot contain colon`)
	})

	t.Run("send request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := client.New(server.URL)

		err := c.RemoveLink(ctx, id)

		assert.Error(t, err, `failed to remove link: 500, Internal Server Error`)
	})

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := client.New(server.URL)

		err := c.RemoveLink(ctx, id)

		assert.Nil(t, err)
	})
}

func TestTranslateLink(t *testing.T) {
	const (
		badID   = "bad uuid"
		goodID  = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
		badType = "bad type"
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			name  string
			rLink rest.Link
			err   string
		}{
			{
				name: "bad id",
				rLink: rest.Link{
					ID: badID,
				},
				err: "received invalid link ID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad ownerID",
				rLink: rest.Link{
					ID:      goodID,
					OwnerID: badID,
				},
				err: "received invalid link ownerID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad locationID",
				rLink: rest.Link{
					ID:         goodID,
					OwnerID:    goodID,
					LocationID: badID,
				},
				err: "received invalid link locationID: 'bad uuid': invalid UUID length: 8",
			},
			{
				name: "bad destinationID",
				rLink: rest.Link{
					ID:            goodID,
					OwnerID:       goodID,
					LocationID:    goodID,
					DestinationID: badID,
				},
				err: "received invalid link destinationID: 'bad uuid': invalid UUID length: 8",
			},
		}

		for _, test := range tests {
			t.Logf("name: %s", test.name)
			i, err := client.TranslateLink(test.rLink)
			assert.Nil(t, i)
			assert.Error(t, err, test.err)
		}
	})

	t.Run("success", func(t *testing.T) {
		const (
			name        = "name"
			desc        = "desc"
			id          = "db81f22a-90cf-48a7-93a2-94de93a9b48f"
			owner       = "7f5908a2-3f99-4e21-a621-d369cff3b061"
			location    = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
			destination = "5ce8aab9-4c49-4870-98f0-4bba0316727a"
		)
		var (
			linkID        = asset.LinkID(uuid.MustParse(id))
			ownerID       = asset.PlayerID(uuid.MustParse(owner))
			locationID    = asset.RoomID(uuid.MustParse(location))
			destinationID = asset.RoomID(uuid.MustParse(destination))
			created       = asset.Timestamp{Time: time.Now().UTC()}
			updated       = asset.Timestamp{Time: time.Now().UTC()}
		)

		rLink := rest.Link{
			ID:            id,
			Name:          name,
			Description:   desc,
			OwnerID:       owner,
			LocationID:    location,
			DestinationID: destination,
			Created:       created,
			Updated:       updated,
		}
		aLink := &asset.Link{
			ID:            linkID,
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
			Created:       created,
			Updated:       updated,
		}

		p, err := client.TranslateLink(rLink)
		assert.Nil(t, err)
		assert.Compare(t, p, aLink)
	})
}

func TestTranslateLinkChange(t *testing.T) {
	const (
		name        = "name"
		desc        = "desc"
		owner       = "7f5908a2-3f99-4e21-a621-d369cff3b061"
		location    = "a4a4474a-a44e-47f9-9b26-c66daa42f2db"
		destination = "5ce8aab9-4c49-4870-98f0-4bba0316727a"
	)

	var (
		ownerID       = asset.PlayerID(uuid.MustParse(owner))
		locationID    = asset.RoomID(uuid.MustParse(location))
		destinationID = asset.RoomID(uuid.MustParse(destination))
	)

	t.Run("failures", func(t *testing.T) {
		tests := []struct {
			change asset.LinkChange
			err    string
		}{
			{
				change: asset.LinkChange{},
				err:    "attempted to send empty name in request",
			},
			{
				change: asset.LinkChange{
					Name: name,
				},
				err: "attempted to send empty description in request",
			},
		}

		for _, test := range tests {
			_, err := client.TranslateLinkChange(test.change)
			assert.Error(t, err, test.err)
		}
	})

	t.Run("success", func(t *testing.T) {
		change := asset.LinkChange{
			Name:          name,
			Description:   desc,
			OwnerID:       ownerID,
			LocationID:    locationID,
			DestinationID: destinationID,
		}
		req, err := client.TranslateLinkChange(change)
		assert.Nil(t, err)
		assert.Equal(t, req, rest.LinkRequest{
			Name:          name,
			Description:   desc,
			OwnerID:       owner,
			LocationID:    location,
			DestinationID: destination,
		})
	})
}
