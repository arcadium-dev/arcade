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
	// ItemManager coordinates the persistent storage of items persented by the networking layer.
	ItemManager struct {
		Storage ItemStorage
	}

	// ItemStorage defines the interface to manage the persistent storage of items.
	ItemStorage interface {
		List(ctx context.Context, filter arcade.ItemsFilter) ([]*arcade.Item, error)
		Get(ctx context.Context, itemID arcade.ItemID) (*arcade.Item, error)
		Create(ctx context.Context, itemReq arcade.ItemRequest) (*arcade.Item, error)
		Update(ctx context.Context, itemID arcade.ItemID, itemReq arcade.ItemRequest) (*arcade.Item, error)
		Remove(ctx context.Context, itemID arcade.ItemID) error
	}
)

// List returns a slice of items based on the value of the filter.
func (m ItemManager) List(ctx context.Context, filter arcade.ItemsFilter) ([]*arcade.Item, error) {
	return m.Storage.List(ctx, filter)
}

// Get returns a single item given the itemID.
func (m ItemManager) Get(ctx context.Context, itemID arcade.ItemID) (*arcade.Item, error) {
	return m.Storage.Get(ctx, itemID)
}

// Create creates a new item in persistent storage.
func (m ItemManager) Create(ctx context.Context, itemReq arcade.ItemRequest) (*arcade.Item, error) {
	return m.Storage.Create(ctx, itemReq)
}

// Update replaces the item in persistent storage.
func (m ItemManager) Update(ctx context.Context, itemID arcade.ItemID, itemReq arcade.ItemRequest) (*arcade.Item, error) {
	return m.Storage.Update(ctx, itemID, itemReq)
}

// Remove deletes the given item, based on the given itemID, from persistent storage.
func (m ItemManager) Remove(ctx context.Context, itemID arcade.ItemID) error {
	return m.Storage.Remove(ctx, itemID)
}
