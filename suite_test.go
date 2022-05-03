package arcade_test

import (
	"math/rand"
	"sync"
	"time"
)

func randString(size int) string {
	var once sync.Once
	once.Do(func() {
		rand.Seed(time.Now().Unix())
	})

	s := "abcdefghijklmnopqrstuvwxyz "
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = s[rand.Intn(len(s))]
	}
	return string(b)
}
