package assets_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"arcadium.dev/core/assert"

	"arcadium.dev/arcade/assets"
)

type Thing struct {
	T assets.Timestamp
}

func TestTimestampMarshalJSON(t *testing.T) {
	s := "2023-09-25T20:10:00.123456"

	t.Run("standalone", func(t *testing.T) {
		ti, err := time.Parse(assets.TimestampFormat, s)
		assert.Nil(t, err)

		ts := assets.Timestamp{Time: ti}
		b, err := ts.MarshalJSON()

		assert.Nil(t, err)
		assert.Equal(t, string(b), fmt.Sprintf("\"%s\"", s))
	})

	t.Run("embedded", func(t *testing.T) {
		ti, err := time.Parse(assets.TimestampFormat, s)
		assert.Nil(t, err)

		thing := Thing{T: assets.Timestamp{Time: ti}}
		b, err := json.Marshal(thing)

		assert.Nil(t, err)
		assert.Equal(t, string(b), `{"T":"2023-09-25T20:10:00.123456"}`)
	})
}

func TestTimestampUnmarshalJSON(t *testing.T) {
	s := "2023-09-25T20:10:00.123456"

	t.Run("error", func(t *testing.T) {
		var ts assets.Timestamp

		err := ts.UnmarshalJSON([]byte(""))
		assert.Error(t, err, `failed to unmarshal timestamp, invalid timestamp`)

		err = ts.UnmarshalJSON([]byte("4"))
		assert.Error(t, err, `failed to unmarshal timestamp, invalid timestamp`)

		err = ts.UnmarshalJSON([]byte(s))
		assert.Error(t, err, `failed to unmarshal timestamp, invalid timestamp`)
	})

	t.Run("standalone", func(t *testing.T) {
		var ts assets.Timestamp
		err := ts.UnmarshalJSON([]byte("\"" + s + "\""))
		assert.Nil(t, err)

		got := ts.Format(assets.TimestampFormat)
		want := s
		assert.Equal(t, got, want)
	})

	t.Run("embedded", func(t *testing.T) {
		b := []byte(`{"T":"` + s + `"}`)
		var got Thing
		assert.Nil(t, json.Unmarshal(b, &got))

		ti, err := time.Parse(assets.TimestampFormat, s)
		assert.Nil(t, err)
		want := Thing{T: assets.Timestamp{Time: ti}}

		assert.Equal(t, got, want)
	})
}
