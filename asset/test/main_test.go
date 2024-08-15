package integration_test

import (
	"math/rand"
	"os"
	"testing"

	"github.com/google/uuid"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest/client"
)

var (
	assets *client.Client

	nothing = asset.ItemID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	nobody  = asset.PlayerID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	nowhere = asset.RoomID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
)

func TestMain(m *testing.M) {
	assets = client.New("https://localhost:4210", client.WithInsecure())
	os.Exit(m.Run())
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
