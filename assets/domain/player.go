//  Copyright 2022-2023 arcadium.dev <info@arcadium.dev>
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

package domain // import "arcadium.dev/arcade/assets/domain"

import (
	"context"

	"arcadium.dev/arcade/assets"
)

type (
	// PlayerManager coordinates the persistent storage of plaeyers persented by the networking layer.
	PlayerManager struct {
		Storage PlayerStorage
	}

	// PlayerStorage defines the interface to manage the persistent storage of items.
	PlayerStorage interface {
		List(context.Context, assets.PlayerFilter) ([]*assets.Player, error)
		Get(context.Context, assets.PlayerID) (*assets.Player, error)
		Create(context.Context, assets.PlayerCreate) (*assets.Player, error)
		Update(context.Context, assets.PlayerID, assets.PlayerUpdate) (*assets.Player, error)
		Remove(context.Context, assets.PlayerID) error
	}
)

// List returns a slice of players based on the value of the filter.
func (m PlayerManager) List(ctx context.Context, filter assets.PlayerFilter) ([]*assets.Player, error) {
	return m.Storage.List(ctx, filter)
}

// Get returns a single player given the playerID.
func (m PlayerManager) Get(ctx context.Context, id assets.PlayerID) (*assets.Player, error) {
	return m.Storage.Get(ctx, id)
}

// Create creates a new player in persistent storage.
func (m PlayerManager) Create(ctx context.Context, create assets.PlayerCreate) (*assets.Player, error) {
	return m.Storage.Create(ctx, create)
}

// Update replaces the player in persistent storage.
func (m PlayerManager) Update(ctx context.Context, id assets.PlayerID, update assets.PlayerUpdate) (*assets.Player, error) {
	return m.Storage.Update(ctx, id, update)
}

// Remove deletes the given player, based on the given playerID, from persistent storage.
func (m PlayerManager) Remove(ctx context.Context, id assets.PlayerID) error {
	return m.Storage.Remove(ctx, id)
}
