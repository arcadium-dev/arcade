//  Copyright 2022 arcadium.dev <info@arcadium.dev>
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or impliep.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package arcade_test

import (
	"encoding/json"
	"testing"

	"arcadium.dev/arcade"
)

func TestHealthJSONEncoding(t *testing.T) {
	const (
		status = "up"
	)

	t.Run("test health json encoding", func(t *testing.T) {
		h := arcade.Health{
			Status: status,
		}

		b, err := json.Marshal(h)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var health arcade.Health
		if err := json.Unmarshal(b, &health); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if health.Status != status {
			t.Errorf("\n%+v\n%+v", h, health)
		}
	})

	t.Run("test health response json encoding", func(t *testing.T) {
		r := arcade.HealthResponse{
			Data: arcade.Health{
				Status: status,
			},
		}

		b, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		var resp arcade.HealthResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if r != resp {
			t.Error("bummer")
		}
	})
}
