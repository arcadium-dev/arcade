package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"arcadium.dev/arcade/asset/rest/client"
	"arcadium.dev/core/assert"
)

func TestClientSend(t *testing.T) {
	ctx := context.Background()

	t.Run("response error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(client.ResponseError{
				Status: http.StatusBadRequest,
				Detail: "response error detail",
			})
			assert.Nil(t, err)
		}))
		defer server.Close()

		c := client.New(server.URL)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/testing", nil)
		assert.Nil(t, err)

		_, err = c.Send(ctx, req)
		assert.Error(t, err, "response error detail")
	})

	t.Run("no response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		c := client.New(server.URL)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/testing", nil)
		assert.Nil(t, err)

		_, err = c.Send(ctx, req)
		assert.Error(t, err, "400, Bad Request")
	})

}
