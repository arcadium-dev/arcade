package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/require"

	"arcadium.dev/arcade/asset"
)

func TestAssets(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration tests: set INTEGRATION environment variable")
	}

	var (
		ctx = context.Background()

		players = make(map[string]*asset.Player)
		homes   = make(map[string]*asset.Room)
		beds    = make(map[string]*asset.Item)
		outs    = make(map[string]*asset.Link)
		ins     = make(map[string]*asset.Link)

		outside *asset.Room
	)

	t.Run("create stuff", func(t *testing.T) {
		var err error

		outside, err = assets.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        "Outside",
				Description: "It's a beautiful day outside.",
				OwnerID:     nobody,
				ParentID:    nowhere,
			},
		})
		assert.Nil(t, err)

		for i := 0; i < 10; i++ {
			name := randName(8)

			player, err := assets.CreatePlayer(ctx, asset.PlayerCreate{
				PlayerChange: asset.PlayerChange{
					Name:        name,
					Description: fmt.Sprintf("This is %s", name),
					HomeID:      nowhere,
					LocationID:  nowhere,
				},
			})
			assert.Nil(t, err)
			players[player.Name] = player

			home, err := assets.CreateRoom(ctx, asset.RoomCreate{
				RoomChange: asset.RoomChange{
					Name:        fmt.Sprintf("%s's Home", player.Name),
					Description: fmt.Sprintf("This is %s's Home", player.Name),
					OwnerID:     player.ID,
					ParentID:    outside.ID,
				},
			})
			assert.Nil(t, err)
			homes[player.Name] = home

			player, err = assets.UpdatePlayer(ctx, player.ID, asset.PlayerUpdate{
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

			bed, err := assets.CreateItem(ctx, asset.ItemCreate{
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

			out, err := assets.CreateLink(ctx, asset.LinkCreate{
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

			in, err := assets.CreateLink(ctx, asset.LinkCreate{
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
		p, err := assets.ListPlayers(ctx, asset.PlayerFilter{Limit: 100})

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
			p, err := assets.ListPlayers(ctx, asset.PlayerFilter{
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
		p, err := assets.ListPlayers(ctx, asset.PlayerFilter{
			LocationID: outside.ID,
		})

		assert.Nil(t, err)
		assert.Equal(t, len(p), 0)
	})

	t.Run("get players", func(t *testing.T) {
		for _, player := range players {
			p, err := assets.GetPlayer(ctx, player.ID)

			assert.Nil(t, err)
			assert.Equal(t, *p, *player)
		}
	})

	t.Run("list homes/rooms", func(t *testing.T) {
		r, err := assets.ListRooms(ctx, asset.RoomFilter{Limit: 100})

		assert.Nil(t, err)
		assert.Equal(t, len(r), len(homes)+2) // including outside and nowhere
	})

	t.Run("list rooms in outside", func(t *testing.T) {
		r, err := assets.ListRooms(ctx, asset.RoomFilter{ParentID: outside.ID})

		assert.Nil(t, err)
		assert.Equal(t, len(r), len(homes))
	})

	t.Run("list rooms in nowhere", func(t *testing.T) {
		r, err := assets.ListRooms(ctx, asset.RoomFilter{ParentID: nowhere})

		t.Logf("nowhere rooms: %+v", r)
		for _, room := range r {
			t.Logf("room: %s", room.Name)
		}

		assert.Nil(t, err)
		require.Equal(t, len(r), 2) // nowhere (is it's own parent) and outside
	})

	t.Run("list rooms owned by each player", func(t *testing.T) {
		for _, player := range players {
			r, err := assets.ListRooms(ctx, asset.RoomFilter{OwnerID: player.ID})

			assert.Nil(t, err)
			require.Equal(t, len(r), 1)
			assert.Equal(t, r[0].ID, homes[player.Name].ID)
		}
	})

	t.Run("get each home", func(t *testing.T) {
		for _, home := range homes {
			r, err := assets.GetRoom(ctx, home.ID)

			assert.Nil(t, err)
			assert.Equal(t, *r, *home)
		}
	})

	t.Run("list beds/items", func(t *testing.T) {
		i, err := assets.ListItems(ctx, asset.ItemFilter{Limit: 100})

		assert.Nil(t, err)
		assert.Equal(t, len(i), len(beds)+1) // including nothing
	})

	t.Run("list beds/items owned by each player", func(t *testing.T) {
		for _, player := range players {
			i, err := assets.ListItems(ctx, asset.ItemFilter{OwnerID: player.ID})

			assert.Nil(t, err)
			require.Equal(t, len(i), 1)
			item := i[0]
			assert.Equal(t, *item, *(beds[player.Name]))
		}
	})

	t.Run("list beds/items located in each room", func(t *testing.T) {
		for _, home := range homes {
			i, err := assets.ListItems(ctx, asset.ItemFilter{LocationID: home.ID})

			assert.Nil(t, err)
			require.Equal(t, len(i), 1)
			item := i[0]
			assert.Equal(t, item.OwnerID, home.OwnerID)
		}
	})

	t.Run("list items in nowhere, nobody and nothing", func(t *testing.T) {
		i, err := assets.ListItems(ctx, asset.ItemFilter{LocationID: nowhere})

		assert.Nil(t, err)
		assert.Equal(t, len(i), 1)
		assert.Equal(t, i[0].ID, nothing)

		i, err = assets.ListItems(ctx, asset.ItemFilter{LocationID: nobody})

		assert.Nil(t, err)
		assert.Equal(t, len(i), 0)

		i, err = assets.ListItems(ctx, asset.ItemFilter{LocationID: nothing})

		assert.Nil(t, err)
		assert.Equal(t, len(i), 0)
	})

	t.Run("list links located outside", func(t *testing.T) {
		l, err := assets.ListLinks(ctx, asset.LinkFilter{LocationID: outside.ID})

		assert.Nil(t, err)
		assert.Equal(t, len(l), len(ins))
	})

	t.Run("list links in each home", func(t *testing.T) {
		var links []*asset.Link
		for _, home := range homes {
			l, err := assets.ListLinks(ctx, asset.LinkFilter{LocationID: home.ID})

			assert.Nil(t, err)
			assert.Equal(t, len(l), 1)
			links = append(links, l...)
		}
		assert.Equal(t, len(links), len(outs))
	})

	t.Run("delete stuff", func(t *testing.T) {
		for _, player := range players {
			assert.Nil(t, assets.RemoveLink(ctx, ins[player.Name].ID))
			assert.Nil(t, assets.RemoveLink(ctx, outs[player.Name].ID))
			assert.Nil(t, assets.RemoveItem(ctx, beds[player.Name].ID))
			assert.Nil(t, assets.RemoveRoom(ctx, homes[player.Name].ID))
			assert.Nil(t, assets.RemovePlayer(ctx, player.ID))
		}
		assert.Nil(t, assets.RemoveRoom(ctx, outside.ID))
	})
}

func TestDeletingStuff(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration tests: set INTEGRATION environment variable")
	}

	t.Run("delete player home", func(t *testing.T) {
	})

	t.Run("delete player location", func(t *testing.T) {
	})

	t.Run("delete room owner", func(t *testing.T) {
	})

	t.Run("delete room parent", func(t *testing.T) {
	})

	t.Run("delete item owner", func(t *testing.T) {
	})

	t.Run("delete item location", func(t *testing.T) {
	})

	t.Run("delete link owner", func(t *testing.T) {
	})

	t.Run("delete link location", func(t *testing.T) {
	})

	t.Run("delete link destination", func(t *testing.T) {
	})
}
