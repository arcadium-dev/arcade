package asset_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"arcadium.dev/core/assert"

	"arcadium.dev/arcade/asset"
)

type Thing struct {
	T asset.Timestamp
}

func TestTimestampMarshalJSON(t *testing.T) {
	s := "2023-09-25T20:10:00.123456"

	t.Run("standalone", func(t *testing.T) {
		ti, err := time.Parse(asset.TimestampFormat, s)
		assert.Nil(t, err)

		ts := asset.Timestamp{Time: ti}
		b, err := ts.MarshalJSON()

		assert.Nil(t, err)
		assert.Equal(t, string(b), fmt.Sprintf("\"%s\"", s))
	})

	t.Run("embedded", func(t *testing.T) {
		ti, err := time.Parse(asset.TimestampFormat, s)
		assert.Nil(t, err)

		thing := Thing{T: asset.Timestamp{Time: ti}}
		b, err := json.Marshal(thing)

		assert.Nil(t, err)
		assert.Equal(t, string(b), `{"T":"2023-09-25T20:10:00.123456"}`)
	})
}

func TestTimestampUnmarshalJSON(t *testing.T) {
	s := "2023-09-25T20:10:00.123456"

	t.Run("error", func(t *testing.T) {
		var ts asset.Timestamp

		err := ts.UnmarshalJSON([]byte(""))
		assert.Error(t, err, `failed to unmarshal timestamp, invalid timestamp`)

		err = ts.UnmarshalJSON([]byte("4"))
		assert.Error(t, err, `failed to unmarshal timestamp, invalid timestamp`)

		err = ts.UnmarshalJSON([]byte(s))
		assert.Error(t, err, `failed to unmarshal timestamp, invalid timestamp`)
	})

	t.Run("standalone", func(t *testing.T) {
		var ts asset.Timestamp
		err := ts.UnmarshalJSON([]byte("\"" + s + "\""))
		assert.Nil(t, err)

		got := ts.Format(asset.TimestampFormat)
		want := s
		assert.Equal(t, got, want)
	})

	t.Run("embedded", func(t *testing.T) {
		b := []byte(`{"T":"` + s + `"}`)
		var got Thing
		assert.Nil(t, json.Unmarshal(b, &got))

		ti, err := time.Parse(asset.TimestampFormat, s)
		assert.Nil(t, err)
		want := Thing{T: asset.Timestamp{Time: ti}}

		assert.Equal(t, got, want)
	})
}
