package postgres_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/require"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/data"
	"arcadium.dev/arcade/data/postgres"
	"arcadium.dev/arcade/user"
)

func randLogin(size int) string {
	s := "abcdefghijklmnopqrstuvwxyz"
	ls := len(s)
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = s[rand.Intn(ls)]
	}
	return string(b)
}

func randPublicKey(login string) []byte {
	s := "ABCDEFGHIJKLMOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz012345789+"
	ls := len(s)
	key := make([]byte, 64)
	for i := 0; i < 64; i++ {
		key[i] = s[rand.Intn(ls)]
	}
	return []byte(fmt.Sprintf("ssh-ed25519 AAAAC%s %s@example.com", key, login))
}

func TestUserDriver(t *testing.T) {
	var (
		ctx = context.Background()

		users   = make(map[string]*user.User)
		players = make(map[string]*asset.Player)
	)

	userStore := data.UserStorage{
		DB: db,
		Driver: postgres.UserDriver{
			Driver: postgres.Driver{},
		},
	}

	playerStore := data.PlayerStorage{
		DB: db,
		Driver: postgres.PlayerDriver{
			Driver: postgres.Driver{},
		},
	}

	var err error

	for i := 0; i < 50; i++ {
		login := randLogin(8)
		pubKey1 := randPublicKey(login)
		pubKey2 := randPublicKey(login)

		u1, err := userStore.Create(ctx, user.Create{
			Change: user.Change{
				Login:     login,
				PublicKey: pubKey1,
			},
		})
		require.Nil(t, err)
		users[u1.Login] = u1

		u2, err := userStore.Update(ctx, u1.ID, user.Update{
			Change: user.Change{
				Login:     u1.Login,
				PublicKey: pubKey2,
			},
		})
		require.Nil(t, err)
		assert.Equal(t, u1.Login, u2.Login)
		assert.Compare(t, u2.PublicKey, pubKey2)

		name := randName(8)
		player, err := playerStore.Create(ctx, asset.PlayerCreate{
			PlayerChange: asset.PlayerChange{
				Name:        name,
				Description: fmt.Sprintf("This is %s", name),
				HomeID:      nowhere,
				LocationID:  nowhere,
			},
		})
		require.Nil(t, err)
		players[player.Name] = player

		u3, err := userStore.AssociatePlayer(ctx, u1.ID, user.AssociatePlayer{PlayerID: player.ID})
		require.Nil(t, err)
		assert.Equal(t, u1.Login, u3.Login)
		assert.Compare(t, u2.PublicKey, u3.PublicKey)
		assert.Compare(t, u3.PlayerID, player.ID)
	}

	playerList, err := playerStore.List(ctx, asset.PlayerFilter{})
	require.Nil(t, err)
	assert.Equal(t, len(playerList)-1, len(players)) // account for nobody

	for _, player := range players {
		if player.ID == nobody {
			continue
		}
		err = playerStore.Remove(ctx, player.ID)
		assert.Nil(t, err)
	}

	userList, err := userStore.List(ctx, user.Filter{})
	require.Nil(t, err)
	assert.Equal(t, len(userList), len(users))

	for _, u := range userList {
		assert.Equal(t, u.PlayerID, nobody)

		err := userStore.Remove(ctx, u.ID)
		assert.Nil(t, err)
	}

	userList, err = userStore.List(ctx, user.Filter{})
	require.Nil(t, err)
	assert.Equal(t, len(userList), 0)
}
