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
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package players

import "testing"

func TestHandlerList(t *testing.T) {
	t.Run("service: unknown error", func(t *testing.T) {
	})

	t.Run("success", func(t *testing.T) {
	})
}

func TestHandlerGet(t *testing.T) {
	t.Run("service: invalid player id", func(t *testing.T) {
	})

	t.Run("service: not found", func(t *testing.T) {
	})

	t.Run("service: unknown error", func(t *testing.T) {
	})

	t.Run("success", func(t *testing.T) {
	})
}

func TestHandlerCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
	})
}

func TestHandlerUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
	})
}

func TestHandlerUpsert(t *testing.T) {
	t.Run("handler: missing body", func(t *testing.T) {
	})

	t.Run("handler: invalid json", func(t *testing.T) {
	})

	t.Run("service: invalid argument", func(t *testing.T) {
	})

	t.Run("service: already exists", func(t *testing.T) {
	})

	t.Run("service: unknown error", func(t *testing.T) {
	})

	t.Run("success", func(t *testing.T) {
	})
}

func TestHandlerRemove(t *testing.T) {
	t.Run("service: invalid argument", func(t *testing.T) {
	})

	t.Run("service: not found", func(t *testing.T) {
	})

	t.Run("service: unknown error", func(t *testing.T) {
	})

	t.Run("success", func(t *testing.T) {
	})
}
