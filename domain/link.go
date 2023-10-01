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
	// LinkManager coordinates the persistent storage of links persented by the networking layer.
	LinkManager struct {
		Storage LinkStorage
	}

	// LinkStorage defines the interface to manage the persistent storage of items.
	LinkStorage interface {
		List(ctx context.Context, filter arcade.LinksFilter) ([]*arcade.Link, error)
		Get(ctx context.Context, linkID arcade.LinkID) (*arcade.Link, error)
		Create(ctx context.Context, link arcade.Link) error
		Update(ctx context.Context, link arcade.Link) (*arcade.Link, error)
		Remove(ctx context.Context, linkID arcade.LinkID) error
	}
)

// List returns a slice of links based on the value of the filter.
func (m LinkManager) List(ctx context.Context, filter arcade.LinksFilter) ([]*arcade.Link, error) {
	return m.Storage.List(ctx, filter)
}

// Get returns a single link given the linkID.
func (m LinkManager) Get(ctx context.Context, linkID arcade.LinkID) (*arcade.Link, error) {
	return m.Storage.Get(ctx, linkID)
}

// Create creates a new link in persistent storage.
func (m LinkManager) Create(ctx context.Context, link arcade.Link) error {
	return m.Storage.Create(ctx, link)
}

// Update replaces the link in persistent storage.
func (m LinkManager) Update(ctx context.Context, link arcade.Link) (*arcade.Link, error) {
	return m.Storage.Update(ctx, link)
}

// Remove deletes the given link, based on the given linkID, from persistent storage.
func (m LinkManager) Remove(ctx context.Context, linkID arcade.LinkID) error {
	return m.Storage.Remove(ctx, linkID)
}
