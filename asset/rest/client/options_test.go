package client

import (
	"testing"
	"time"

	"arcadium.dev/core/assert"
)

func TestWithTimeout(t *testing.T) {
	t.Run("invalid value", func(t *testing.T) {
		client := New("", WithTimeout(time.Duration(-1*time.Second)))
		assert.Equal(t, client.client.Timeout, defaultTimeout)
	})

	t.Run("success", func(t *testing.T) {
		timeout := 45 * time.Second
		client := New("", WithTimeout(timeout))
		assert.Equal(t, client.client.Timeout, timeout)
	})
}

/*
func TestWithTLSConfig(t *testing.T) {
		client := New("", WithTLSConfig(&tls.Config{ InsecureSkipVerify: true})
		assert.Equal(t, client.client.Timeout, defaultTimeout)
}
*/
