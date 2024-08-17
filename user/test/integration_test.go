package integration_test

import (
	"context"
	"fmt"
	"testing"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/user"
	"arcadium.dev/core/assert"
	"arcadium.dev/core/require"
)

func TestUsers(t *testing.T) {
	var (
		ctx = context.Background()

		us = make(map[string]*user.User)
		ps = make(map[string]*asset.Player)
	)

	var err error

	for i := 0; i < 20; i++ {
		login := randLogin(8)
		pubKey1 := randPublicKey(login)
		pubKey2 := randPublicKey(login)

		u1, err := users.Create(ctx, user.Create{
			Change: user.Change{
				Login:     login,
				PublicKey: pubKey1,
			},
		})
		require.Nil(t, err)
		us[u1.Login] = u1

		u2, err := users.Update(ctx, u1.ID, user.Update{
			Change: user.Change{
				Login:     u1.Login,
				PublicKey: pubKey2,
			},
		})
		require.Nil(t, err)
		assert.Equal(t, u1.Login, u2.Login)
		assert.Compare(t, u2.PublicKey, pubKey2)

		name := randName(8)
		player, err := assets.CreatePlayer(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: fmt.Sprintf("This is %s", name),
				HomeID:      nowhere,
				LocationID:  nowhere,
			},
		})
		require.Nil(t, err)
		ps[player.Name] = player

		u3, err := users.AssociatePlayer(ctx, u1.ID, user.AssociatePlayer{PlayerID: player.ID})
		require.Nil(t, err)
		assert.Equal(t, u1.Login, u3.Login)
		assert.Compare(t, u2.PublicKey, u3.PublicKey)
		assert.Compare(t, u3.PlayerID, player.ID)
	}

	playerList, err := assets.ListPlayers(ctx, asset.PlayerFilter{})
	require.Nil(t, err)
	assert.Equal(t, len(playerList)-1, len(ps)) // account for nobody

	for _, player := range ps {
		if player.ID == nobody {
			continue
		}
		err = assets.RemovePlayer(ctx, player.ID)
		assert.Nil(t, err)
	}

	userList, err := users.List(ctx, user.Filter{})
	require.Nil(t, err)
	assert.Equal(t, len(userList), len(us))

	for _, u := range userList {
		assert.Equal(t, u.PlayerID, nobody)

		err := users.Remove(ctx, u.ID)
		assert.Nil(t, err)
	}

	userList, err = users.List(ctx, user.Filter{})
	require.Nil(t, err)
	assert.Equal(t, len(userList), 0)
}
