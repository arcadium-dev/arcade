package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/errors"
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

	t.Run("create resources", func(t *testing.T) {
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

	t.Run("get items", func(t *testing.T) {
		for _, bed := range beds {
			i, err := assets.GetItem(ctx, bed.ID)

			assert.Nil(t, err)
			assert.Equal(t, *i, *bed)
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

	t.Run("get each link", func(t *testing.T) {
		for _, in := range ins {
			l, err := assets.GetLink(ctx, in.ID)

			assert.Nil(t, err)
			assert.Equal(t, *l, *in)
		}
		for _, out := range outs {
			l, err := assets.GetLink(ctx, out.ID)

			assert.Nil(t, err)
			assert.Equal(t, *l, *out)
		}
	})

	t.Run("delete resources", func(t *testing.T) {
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

func TestOnDeleteSetDefault(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration tests: set INTEGRATION environment variable")
	}

	var (
		ctx = context.Background()
	)

	t.Run("delete player home", func(t *testing.T) {
		home, err := assets.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        "Welcome home!",
				Description: "Home sweet home",
				OwnerID:     nobody,
				ParentID:    nowhere,
			},
		})
		assert.Nil(t, err)

		name := randName(8)
		player, err := assets.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: fmt.Sprintf("This is %s", name),
				HomeID:      home.ID,
				LocationID:  home.ID,
			},
		})
		assert.Nil(t, err)

		err = assets.RemoveRoom(ctx, home.ID)
		assert.Nil(t, err)

		player, err = assets.GetPlayer(ctx, player.ID)
		assert.Nil(t, err)
		assert.Equal(t, player.HomeID, nowhere)
		assert.Equal(t, player.LocationID, nowhere)

		err = assets.RemovePlayer(ctx, player.ID)
		assert.Nil(t, err)

		_, err = assets.GetPlayer(ctx, player.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})

	t.Run("delete room owner", func(t *testing.T) {
		player, err := assets.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        randName(8),
				Description: "A player",
				HomeID:      nowhere,
				LocationID:  nowhere,
			},
		})

		assert.Nil(t, err)
		room, err := assets.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        "A room",
				Description: "This is a room.",
				OwnerID:     player.ID,
				ParentID:    nowhere,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, room.OwnerID, player.ID)

		err = assets.RemovePlayer(ctx, player.ID)
		assert.Nil(t, err)

		room, err = assets.GetRoom(ctx, room.ID)
		assert.Nil(t, err)
		assert.Equal(t, room.OwnerID, nobody)

		err = assets.RemoveRoom(ctx, room.ID)
		assert.Nil(t, err)

		_, err = assets.GetRoom(ctx, room.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})

	t.Run("delete room parent", func(t *testing.T) {
		world, err := assets.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        "World",
				Description: "This is the world.",
				OwnerID:     nobody,
				ParentID:    nowhere,
			},
		})
		assert.Nil(t, err)

		room, err := assets.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        "A room",
				Description: "This is a room",
				OwnerID:     nobody,
				ParentID:    world.ID,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, room.ParentID, world.ID)

		err = assets.RemoveRoom(ctx, world.ID)
		assert.Nil(t, err)

		room, err = assets.GetRoom(ctx, room.ID)
		assert.Nil(t, err)
		assert.Equal(t, room.ParentID, nowhere)

		err = assets.RemoveRoom(ctx, room.ID)
		assert.Nil(t, err)

		_, err = assets.GetRoom(ctx, room.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})

	t.Run("delete item owner", func(t *testing.T) {
		player, err := assets.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        randName(8),
				Description: "A player",
				HomeID:      nowhere,
				LocationID:  nowhere,
			},
		})
		assert.Nil(t, err)

		item, err := assets.CreateItem(ctx, asset.ItemCreate{
			ItemChange: asset.ItemChange{
				Name:        "A thing",
				Description: "A nice thing.",
				OwnerID:     player.ID,
				LocationID:  nowhere,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, item.OwnerID, player.ID)

		err = assets.RemovePlayer(ctx, player.ID)
		assert.Nil(t, err)

		item, err = assets.GetItem(ctx, item.ID)
		assert.Nil(t, err)
		assert.Equal(t, item.OwnerID, nobody)

		err = assets.RemoveItem(ctx, item.ID)
		assert.Nil(t, err)

		_, err = assets.GetItem(ctx, item.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})

	t.Run("delete item location - room", func(t *testing.T) {
		room, err := assets.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        "A room",
				Description: "This is a room",
				OwnerID:     nobody,
				ParentID:    nowhere,
			},
		})
		assert.Nil(t, err)

		item, err := assets.CreateItem(ctx, asset.ItemCreate{
			ItemChange: asset.ItemChange{
				Name:        "A thing",
				Description: "A nice thing.",
				OwnerID:     nobody,
				LocationID:  room.ID,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, item.LocationID.ID(), asset.LocationID(room.ID))
		assert.Equal(t, item.LocationID.Type(), asset.LocationTypeRoom)

		err = assets.RemoveRoom(ctx, room.ID)
		assert.Nil(t, err)

		item, err = assets.GetItem(ctx, item.ID)
		assert.Nil(t, err)
		assert.Equal(t, item.LocationID.ID(), asset.LocationID(nowhere))
		assert.Equal(t, item.LocationID.Type(), asset.LocationTypeRoom)

		err = assets.RemoveItem(ctx, item.ID)
		assert.Nil(t, err)

		_, err = assets.GetItem(ctx, item.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})

	t.Run("delete item location - player", func(t *testing.T) {
		player, err := assets.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        randName(8),
				Description: "A player",
				HomeID:      nowhere,
				LocationID:  nowhere,
			},
		})
		assert.Nil(t, err)

		item, err := assets.CreateItem(ctx, asset.ItemCreate{
			ItemChange: asset.ItemChange{
				Name:        "A thing",
				Description: "A nice thing.",
				OwnerID:     nobody,
				LocationID:  player.ID,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, item.LocationID.ID(), asset.LocationID(player.ID))
		assert.Equal(t, item.LocationID.Type(), asset.LocationTypePlayer)

		err = assets.RemovePlayer(ctx, player.ID)
		assert.Nil(t, err)

		item, err = assets.GetItem(ctx, item.ID)
		assert.Nil(t, err)
		assert.Equal(t, item.LocationID.ID(), asset.LocationID(nobody))
		assert.Equal(t, item.LocationID.Type(), asset.LocationTypePlayer)

		err = assets.RemoveItem(ctx, item.ID)
		assert.Nil(t, err)

		_, err = assets.GetItem(ctx, item.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})

	t.Run("delete item location - item", func(t *testing.T) {
		bag, err := assets.CreateItem(ctx, asset.ItemCreate{
			ItemChange: asset.ItemChange{
				Name:        "A bag",
				Description: "A spacious thing.",
				OwnerID:     nobody,
				LocationID:  nowhere,
			},
		})
		assert.Nil(t, err)

		item, err := assets.CreateItem(ctx, asset.ItemCreate{
			ItemChange: asset.ItemChange{
				Name:        "A thing",
				Description: "A nice thing.",
				OwnerID:     nobody,
				LocationID:  bag.ID,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, item.LocationID.ID(), asset.LocationID(bag.ID))
		assert.Equal(t, item.LocationID.Type(), asset.LocationTypeItem)

		err = assets.RemoveItem(ctx, bag.ID)
		assert.Nil(t, err)

		item, err = assets.GetItem(ctx, item.ID)
		assert.Nil(t, err)
		assert.Equal(t, item.LocationID.ID(), asset.LocationID(nothing))
		assert.Equal(t, item.LocationID.Type(), asset.LocationTypeItem)

		err = assets.RemoveItem(ctx, item.ID)
		assert.Nil(t, err)

		_, err = assets.GetItem(ctx, item.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})

	t.Run("delete link owner", func(t *testing.T) {
		player, err := assets.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        randName(8),
				Description: "A player",
				HomeID:      nowhere,
				LocationID:  nowhere,
			},
		})
		assert.Nil(t, err)

		link, err := assets.CreateLink(ctx, asset.LinkCreate{
			LinkChange: asset.LinkChange{
				Name:          "out",
				Description:   "The way out.",
				OwnerID:       player.ID,
				LocationID:    nowhere,
				DestinationID: nowhere,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, link.OwnerID, player.ID)

		err = assets.RemovePlayer(ctx, player.ID)
		assert.Nil(t, err)

		link, err = assets.GetLink(ctx, link.ID)
		assert.Nil(t, err)
		assert.Equal(t, link.OwnerID, nobody)

		err = assets.RemoveLink(ctx, link.ID)
		assert.Nil(t, err)

		_, err = assets.GetLink(ctx, link.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})

	t.Run("delete link location and destination", func(t *testing.T) {
		room, err := assets.CreateRoom(ctx, asset.RoomCreate{
			RoomChange: asset.RoomChange{
				Name:        "A room",
				Description: "This is a room",
				OwnerID:     nobody,
				ParentID:    nowhere,
			},
		})
		assert.Nil(t, err)

		link, err := assets.CreateLink(ctx, asset.LinkCreate{
			LinkChange: asset.LinkChange{
				Name:          "out",
				Description:   "The way out.",
				OwnerID:       nobody,
				LocationID:    room.ID,
				DestinationID: room.ID,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, link.LocationID, room.ID)
		assert.Equal(t, link.DestinationID, room.ID)

		err = assets.RemoveRoom(ctx, room.ID)
		assert.Nil(t, err)

		link, err = assets.GetLink(ctx, link.ID)
		assert.Nil(t, err)
		assert.Equal(t, link.LocationID, nowhere)
		assert.Equal(t, link.DestinationID, nowhere)

		err = assets.RemoveLink(ctx, link.ID)
		assert.Nil(t, err)

		_, err = assets.GetLink(ctx, link.ID)
		assert.IsError(t, err, errors.ErrNotFound)
	})
}
