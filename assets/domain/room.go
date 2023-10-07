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
	// RoomManager coordinates the persistent storage of rooms persented by the networking layer.
	RoomManager struct {
		Storage RoomStorage
	}

	// RoomStorage defines the interface to manage the persistent storage of items.
	RoomStorage interface {
		List(context.Context, assets.RoomsFilter) ([]*assets.Room, error)
		Get(context.Context, assets.RoomID) (*assets.Room, error)
		Create(context.Context, assets.RoomCreate) (*assets.Room, error)
		Update(context.Context, assets.RoomID, assets.RoomUpdate) (*assets.Room, error)
		Remove(context.Context, assets.RoomID) error
	}
)

// List returns a slice of rooms based on the value of the filter.
func (m RoomManager) List(ctx context.Context, filter assets.RoomsFilter) ([]*assets.Room, error) {
	return m.Storage.List(ctx, filter)
}

// Get returns a single room given the roomID.
func (m RoomManager) Get(ctx context.Context, id assets.RoomID) (*assets.Room, error) {
	return m.Storage.Get(ctx, id)
}

// Create creates a new room in persistent storage.
func (m RoomManager) Create(ctx context.Context, create assets.RoomCreate) (*assets.Room, error) {
	return m.Storage.Create(ctx, create)
}

// Update replaces the room in persistent storage.
func (m RoomManager) Update(ctx context.Context, id assets.RoomID, update assets.RoomUpdate) (*assets.Room, error) {
	return m.Storage.Update(ctx, id, update)
}

// Remove deletes the given room, based on the given roomID, from persistent storage.
func (m RoomManager) Remove(ctx context.Context, id assets.RoomID) error {
	return m.Storage.Remove(ctx, id)
}
