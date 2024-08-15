package server_test

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http/httptest"
	"testing"

	"arcadium.dev/core/assert"
	cserver "arcadium.dev/core/http/server"
)

func assertRespError(t *testing.T, w *httptest.ResponseRecorder, status int, errMsg string) {
	t.Helper()

	resp := w.Result()
	assert.Equal(t, resp.StatusCode, status)

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	defer resp.Body.Close()

	var respErr cserver.ResponseError
	err = json.Unmarshal(body, &respErr)
	assert.Nil(t, err)

	assert.Contains(t, respErr.Detail, errMsg)
	assert.Equal(t, respErr.Status, status)
}

func randString(size int) string {
	s := "abcdefghijklmnopqrstuvwxyz "
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = s[rand.Intn(len(s))]
	}
	return string(b)
}
