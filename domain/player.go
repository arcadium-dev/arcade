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

package domain // import "arcadium.dev/domain"

import (
	"context"

	"arcadium.dev/arcade"
)

type (
	// PlayerManager coordinates the persistent storage of plaeyers persented by the networking layer.
	PlayerManager struct {
		Storage PlayerStorage
	}

	// PlayerStorage defines the interface to manage the persistent storage of items.
	PlayerStorage interface {
		List(ctx context.Context, filter arcade.PlayersFilter) ([]*arcade.Player, error)
		Get(ctx context.Context, playerID arcade.PlayerID) (*arcade.Player, error)
		Create(ctx context.Context, player arcade.Player) error
		Update(ctx context.Context, player arcade.Player) (*arcade.Player, error)
		Remove(ctx context.Context, playerID arcade.PlayerID) error
	}
)

// List returns a slice of players based on the value of the filter.
func (m PlayerManager) List(ctx context.Context, filter arcade.PlayersFilter) ([]*arcade.Player, error) {
	return m.Storage.List(ctx, filter)
}

// Get returns a single player given the playerID.
func (m PlayerManager) Get(ctx context.Context, playerID arcade.PlayerID) (*arcade.Player, error) {
	return m.Storage.Get(ctx, playerID)
}

// Create creates a new player in persistent storage.
func (m PlayerManager) Create(ctx context.Context, player arcade.Player) error {
	return m.Storage.Create(ctx, player)
}

// Update replaces the player in persistent storage.
func (m PlayerManager) Update(ctx context.Context, player arcade.Player) (*arcade.Player, error) {
	return m.Storage.Update(ctx, player)
}

// Remove deletes the given player, based on the given playerID, from persistent storage.
func (m PlayerManager) Remove(ctx context.Context, playerID arcade.PlayerID) error {
	return m.Storage.Remove(ctx, playerID)
}
