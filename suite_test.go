package arcade_test

import "math/rand"

func randString(size int) string {
	s := "abcdefghijklmnopqrstuvwxyz "
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = s[rand.Intn(len(s))]
	}
	return string(b)
}
