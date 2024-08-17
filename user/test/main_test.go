package integration_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/google/uuid"

	"arcadium.dev/arcade/asset"
	aclient "arcadium.dev/arcade/asset/rest/client"
	"arcadium.dev/arcade/user"
	uclient "arcadium.dev/arcade/user/client"
)

var (
	users  *uclient.UsersClient
	assets *aclient.Client

	ctx = context.Background()

	nobody  = asset.PlayerID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	nowhere = asset.RoomID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
)

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION") == "" {
		log.Print("skipping integration tests: set INTEGRATION environment variable")
		os.Exit(0)
	}

	var err error
	users, err = uclient.New("https://localhost:4220", uclient.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	if err != nil {
		log.Fatal(err)
	}

	assets = aclient.New("https://localhost:4210", aclient.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))

	os.Exit(m.Run())
}

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

func createUsers(count int) ([]*user.User, error) {
	us := make([]*user.User, count)
	for i := 0; i < count; i++ {
		login := randLogin(8)
		u, err := users.Create(ctx, user.Create{
			Change: user.Change{
				Login:     login,
				PublicKey: randPublicKey(login),
			},
		})
		if err != nil {
			return nil, err
		}
		us[i] = u
	}
	return us, nil
}
