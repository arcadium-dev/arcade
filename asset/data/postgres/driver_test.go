package postgres_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/require"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/data"
	"arcadium.dev/arcade/asset/data/postgres"
)

var (
	nothing = asset.ItemID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	nobody  = asset.PlayerID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	nowhere = asset.RoomID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
)

func randName(size int) string {
	caps := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	s := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		if i == 0 {
			b[i] = caps[rand.Intn(len(caps))]
			continue
		}
		b[i] = s[rand.Intn(len(s))]
	}
	return string(b)
}

func TestDriver(t *testing.T) {
	var (
		ctx = context.Background()

		players = make(map[string]*asset.Player)
		homes   = make(map[string]*asset.Room)
		beds    = make(map[string]*asset.Item)
		outs    = make(map[string]*asset.Link)
		ins     = make(map[string]*asset.Link)

		outside *asset.Room
	)

	itemStore := data.ItemStorage{
		DB: db,
		Driver: postgres.ItemDriver{
			Driver: postgres.Driver{},
		},
	}

	linkStore := data.LinkStorage{
		DB: db,
		Driver: postgres.LinkDriver{
			Driver: postgres.Driver{},
		},
	}

	playerStore := data.PlayerStorage{
		DB: db,
		Driver: postgres.PlayerDriver{
			Driver: postgres.Driver{},
		},
	}

	roomStore := data.RoomStorage{
		DB: db,
		Driver: postgres.RoomDriver{
			Driver: postgres.Driver{},
		},
	}

	t.Run("create resources", func(t *testing.T) {
		var err error
		outside, err = roomStore.Create(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        "Outside",
				Description: "It's a beautiful day outside.",
				OwnerID:     nobody,
				ParentID:    nowhere,
			},
		})
		assert.Nil(t, err)

		for i := 0; i < 90; i++ {
			name := randName(8)

			player, err := playerStore.Create(ctx, asset.PlayerCreate{
				PlayerChange: asset.PlayerChange{
					Name:        name,
					Description: fmt.Sprintf("This is %s", name),
					HomeID:      nowhere,
					LocationID:  nowhere,
				},
			})
			assert.Nil(t, err)
			players[player.Name] = player

			home, err := roomStore.Create(ctx, asset.RoomCreate{
				RoomChange: asset.RoomChange{
					Name:        fmt.Sprintf("%s's Home", player.Name),
					Description: fmt.Sprintf("This is %s's Home", player.Name),
					OwnerID:     player.ID,
					ParentID:    outside.ID,
				},
			})
			assert.Nil(t, err)
			homes[player.Name] = home

			player, err = playerStore.Update(ctx, player.ID, asset.PlayerUpdate{
				PlayerChange: asset.PlayerChange{
					Name:        player.Name,
					Description: player.Description,
					HomeID:      home.ID,
					LocationID:  home.ID,
				},
			})
			assert.Nil(t, err)
			assert.Equal(t, player.HomeID, home.ID)
			assert.Equal(t, player.LocationID, home.ID)
			assert.NotNil(t, players[player.Name])
			players[player.Name] = player

			bed, err := itemStore.Create(ctx, asset.ItemCreate{
				ItemChange: asset.ItemChange{
					Name:        fmt.Sprintf("%s's Bed", player.Name),
					Description: "Soft and bouncy all at the same time",
					OwnerID:     player.ID,
					LocationID:  home.ID,
				},
			})
			assert.Nil(t, err)
			assert.Equal(t, bed.OwnerID, player.ID)
			assert.Equal(t, bed.LocationID.ID(), asset.LocationID(home.ID))
			beds[player.Name] = bed

			out, err := linkStore.Create(ctx, asset.LinkCreate{
				LinkChange: asset.LinkChange{
					Name:          "out",
					Description:   "The way out.",
					OwnerID:       player.ID,
					LocationID:    home.ID,
					DestinationID: outside.ID,
				},
			})
			assert.Nil(t, err)
			assert.Equal(t, out.OwnerID, player.ID)
			assert.Equal(t, out.LocationID, home.ID)
			assert.Equal(t, out.DestinationID, outside.ID)
			outs[player.Name] = out

			in, err := linkStore.Create(ctx, asset.LinkCreate{
				LinkChange: asset.LinkChange{
					Name:          fmt.Sprintf("To %s's home", player.Name),
					Description:   fmt.Sprintf("To %s's home", player.Name),
					OwnerID:       player.ID,
					LocationID:    outside.ID,
					DestinationID: home.ID,
				},
			})
			assert.Nil(t, err)
			assert.Equal(t, in.OwnerID, player.ID)
			assert.Equal(t, in.LocationID, outside.ID)
			assert.Equal(t, in.DestinationID, home.ID)
			ins[player.Name] = in
		}
	})

	t.Run("list players", func(t *testing.T) {
		p, err := playerStore.List(ctx, asset.PlayerFilter{Limit: 100})

		assert.Nil(t, err)
		assert.Equal(t, len(p), len(players)+1) // plus nobody

		for _, player := range p {
			if player.Name == "Nobody" {
				continue
			}
			require.NotNil(t, players[player.Name])
			assert.Equal(t, *player, *(players[player.Name]))
		}
	})

	t.Run("list players by location", func(t *testing.T) {
		for _, player := range players {
			p, err := playerStore.List(ctx, asset.PlayerFilter{
				LocationID: player.LocationID,
			})

			assert.Nil(t, err)
			require.Equal(t, len(p), 1)
			pl := p[0]
			assert.Equal(t, *pl, *player)
			assert.Equal(t, pl.LocationID, pl.HomeID)
		}
	})

	t.Run("list players outside", func(t *testing.T) {
		p, err := playerStore.List(ctx, asset.PlayerFilter{
			LocationID: outside.ID,
		})

		assert.Nil(t, err)
		assert.Equal(t, len(p), 0)
	})

	t.Run("get players", func(t *testing.T) {
		for _, player := range players {
			p, err := playerStore.Get(ctx, player.ID)

			assert.Nil(t, err)
			assert.Equal(t, *p, *player)
		}
	})

	t.Run("list homes/rooms", func(t *testing.T) {
		r, err := roomStore.List(ctx, asset.RoomFilter{Limit: 100})

		assert.Nil(t, err)
		assert.Equal(t, len(r), len(homes)+2) // including outside and nowhere
	})

	t.Run("list rooms in outside", func(t *testing.T) {
		r, err := roomStore.List(ctx, asset.RoomFilter{ParentID: outside.ID})

		assert.Nil(t, err)
		assert.Equal(t, len(r), len(homes))
	})

	t.Run("list rooms in nowhere", func(t *testing.T) {
		r, err := roomStore.List(ctx, asset.RoomFilter{ParentID: nowhere})

		assert.Nil(t, err)
		require.Equal(t, len(r), 2) // nowhere (is it's own parent) and outside
	})

	t.Run("list rooms owned by each player", func(t *testing.T) {
		for _, player := range players {
			r, err := roomStore.List(ctx, asset.RoomFilter{OwnerID: player.ID})

			assert.Nil(t, err)
			require.Equal(t, len(r), 1)
			assert.Equal(t, r[0].ID, homes[player.Name].ID)
		}
	})

	t.Run("get each home", func(t *testing.T) {
		for _, home := range homes {
			r, err := roomStore.Get(ctx, home.ID)

			assert.Nil(t, err)
			assert.Equal(t, *r, *home)
		}
	})

	t.Run("list beds/items", func(t *testing.T) {
		i, err := itemStore.List(ctx, asset.ItemFilter{Limit: 100})

		assert.Nil(t, err)
		assert.Equal(t, len(i), len(beds)+1) // including nothing
	})

	t.Run("list beds/items owned by each player", func(t *testing.T) {
		for _, player := range players {
			i, err := itemStore.List(ctx, asset.ItemFilter{OwnerID: player.ID})

			assert.Nil(t, err)
			require.Equal(t, len(i), 1)
			item := i[0]
			assert.Equal(t, *item, *(beds[player.Name]))
		}
	})

	t.Run("list beds/items located in each room", func(t *testing.T) {
		for _, home := range homes {
			i, err := itemStore.List(ctx, asset.ItemFilter{LocationID: home.ID})

			assert.Nil(t, err)
			require.Equal(t, len(i), 1)
			item := i[0]
			assert.Equal(t, item.OwnerID, home.OwnerID)
		}
	})

	t.Run("get items", func(t *testing.T) {
		for _, bed := range beds {
			i, err := itemStore.Get(ctx, bed.ID)

			assert.Nil(t, err)
			assert.Equal(t, *i, *bed)
		}
	})

	t.Run("list items in nowhere, nobody and nothing", func(t *testing.T) {
		i, err := itemStore.List(ctx, asset.ItemFilter{LocationID: nowhere})

		assert.Nil(t, err)
		assert.Equal(t, len(i), 1)
		assert.Equal(t, i[0].ID, nothing)

		i, err = itemStore.List(ctx, asset.ItemFilter{LocationID: nobody})

		assert.Nil(t, err)
		assert.Equal(t, len(i), 0)

		i, err = itemStore.List(ctx, asset.ItemFilter{LocationID: nothing})

		assert.Nil(t, err)
		assert.Equal(t, len(i), 0)
	})

	t.Run("list links located outside", func(t *testing.T) {
		l, err := linkStore.List(ctx, asset.LinkFilter{LocationID: outside.ID})

		assert.Nil(t, err)
		assert.Equal(t, len(l), len(ins))
	})

	t.Run("list links in each home", func(t *testing.T) {
		var links []*asset.Link
		for _, home := range homes {
			l, err := linkStore.List(ctx, asset.LinkFilter{LocationID: home.ID})

			assert.Nil(t, err)
			assert.Equal(t, len(l), 1)
			links = append(links, l...)
		}
		assert.Equal(t, len(links), len(outs))
	})

	t.Run("get each link", func(t *testing.T) {
		for _, in := range ins {
			l, err := linkStore.Get(ctx, in.ID)

			assert.Nil(t, err)
			assert.Equal(t, *l, *in)
		}
		for _, out := range outs {
			l, err := linkStore.Get(ctx, out.ID)

			assert.Nil(t, err)
			assert.Equal(t, *l, *out)
		}
	})

	t.Run("delete resources", func(t *testing.T) {
		for _, player := range players {
			assert.Nil(t, linkStore.Remove(ctx, ins[player.Name].ID))
			assert.Nil(t, linkStore.Remove(ctx, outs[player.Name].ID))
			assert.Nil(t, itemStore.Remove(ctx, beds[player.Name].ID))
			assert.Nil(t, roomStore.Remove(ctx, homes[player.Name].ID))
			assert.Nil(t, playerStore.Remove(ctx, player.ID))
		}
		assert.Nil(t, roomStore.Remove(ctx, outside.ID))
	})
}
