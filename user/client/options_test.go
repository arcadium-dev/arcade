package client

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"arcadium.dev/core/assert"
	"arcadium.dev/core/require"
)

func TestWithTimeout(t *testing.T) {
	t.Run("invalid value", func(t *testing.T) {
		c, err := New("", WithTimeout(time.Duration(-1*time.Second)))
		assert.Nil(t, err)
		assert.Equal(t, c.httpClient.Timeout, defaultTimeout)
	})

	t.Run("success", func(t *testing.T) {
		timeout := 45 * time.Second
		c, err := New("", WithTimeout(timeout))
		assert.Nil(t, err)
		assert.Equal(t, c.httpClient.Timeout, timeout)
	})
}

func TestWithTLSConfig(t *testing.T) {
	client, err := New("", WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	assert.Nil(t, err)
	require.NotNil(t, client.httpClient.Transport)
	transport, ok := client.httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}
