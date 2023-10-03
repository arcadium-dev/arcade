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
	// RoomManager coordinates the persistent storage of rooms persented by the networking layer.
	RoomManager struct {
		Storage RoomStorage
	}

	// RoomStorage defines the interface to manage the persistent storage of items.
	RoomStorage interface {
		List(ctx context.Context, filter arcade.RoomsFilter) ([]*arcade.Room, error)
		Get(ctx context.Context, roomID arcade.RoomID) (*arcade.Room, error)
		Create(ctx context.Context, ingressRoom arcade.IngressRoom) (*arcade.Room, error)
		Update(ctx context.Context, roomID arcade.RoomID, ingressRoom arcade.IngressRoom) (*arcade.Room, error)
		Remove(ctx context.Context, roomID arcade.RoomID) error
	}
)

// List returns a slice of rooms based on the value of the filter.
func (m RoomManager) List(ctx context.Context, filter arcade.RoomsFilter) ([]*arcade.Room, error) {
	return m.Storage.List(ctx, filter)
}

// Get returns a single room given the roomID.
func (m RoomManager) Get(ctx context.Context, roomID arcade.RoomID) (*arcade.Room, error) {
	return m.Storage.Get(ctx, roomID)
}

// Create creates a new room in persistent storage.
func (m RoomManager) Create(ctx context.Context, ingressRoom arcade.IngressRoom) (*arcade.Room, error) {
	return m.Storage.Create(ctx, ingressRoom)
}

// Update replaces the room in persistent storage.
func (m RoomManager) Update(ctx context.Context, roomID arcade.RoomID, ingressRoom arcade.IngressRoom) (*arcade.Room, error) {
	return m.Storage.Update(ctx, roomID, ingressRoom)
}

// Remove deletes the given room, based on the given roomID, from persistent storage.
func (m RoomManager) Remove(ctx context.Context, roomID arcade.RoomID) error {
	return m.Storage.Remove(ctx, roomID)
}
