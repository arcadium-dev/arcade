package networking_test

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade/networking"
)

func invokeItemsEndpoint(t *testing.T, m mockItemManager, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := networking.ItemsService{Manager: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, b)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	for i := 0; i < len(query); i += 2 {
		q.Add(query[i], query[i+1])
	}
	r.URL.RawQuery = q.Encode()

	router.ServeHTTP(w, r)

	return w
}

func invokeLinksEndpoint(t *testing.T, m mockLinkManager, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := networking.LinksService{Manager: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, b)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	for i := 0; i < len(query); i += 2 {
		q.Add(query[i], query[i+1])
	}
	r.URL.RawQuery = q.Encode()

	router.ServeHTTP(w, r)

	return w
}

func invokePlayersEndpoint(t *testing.T, m mockPlayerManager, method, target string, body []byte, query ...string) *httptest.ResponseRecorder {
	t.Helper()

	if len(query)%2 != 0 {
		t.Fatal("query param problem, must be divible by 2")
	}

	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}

	router := mux.NewRouter()
	s := networking.PlayersService{Manager: m}
	s.Register(router)

	r := httptest.NewRequest(method, target, b)
	w := httptest.NewRecorder()

	q := r.URL.Query()
	for i := 0; i < len(query); i += 2 {
		q.Add(query[i], query[i+1])
	}
	r.URL.RawQuery = q.Encode()

	router.ServeHTTP(w, r)

	return w
}

func assertRespError(t *testing.T, w *httptest.ResponseRecorder, status int, errMsg string) {
	t.Helper()

	resp := w.Result()
	assert.Equal(t, resp.StatusCode, status)

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	defer resp.Body.Close()

	var respErr server.ResponseError
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
