package integration_test

import (
	"context"
	"fmt"
	"testing"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/core/assert"
)

func TestAssets(t *testing.T) {
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

	// List Players: should be 10

	// List players in each home

	// List players outside: should be 0

	// Get each player

	// List rooms: should be 10

	// List rooms owned by each player

	// List rooms in outside: should be 10

	// List rooms in nowhere: should be 0

	// Get each room

	// List Items: should be 10

	// List items owned by each player

	// List items in each room.

	// List items in nowhere, nobody, nothing: each should be 0

	// List links in outside, should be 10

	// List links in each home, should be 1

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
