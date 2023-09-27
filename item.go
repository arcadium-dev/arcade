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

package arcade // import "arcadium.dev/arcade"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/core/errors"
)

const (
	MaxItemNameLen        = 255
	MaxItemDescriptionLen = 4096
)

type (
	// Item is the internal representation of the data related to a item.
	Item struct {
		ID          string    `json:"itemID"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		OwnerID     string    `json:"ownerID"`
		LocationID  string    `json:"locationID"`
		InventoryID string    `json:"inventoryID"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
	}

	// ItemRequest is the payload of a item create or update request.
	ItemRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		OwnerID     string `json:"ownerID"`
		LocationID  string `json:"locationID"`
		InventoryID string `json:"inventoryID"`
	}

	// ItemResponse is used to json encoded a single item response.
	ItemResponse struct {
		Data Item `json:"data"`
	}

	// ItemsResponse is used to json encoded a multi-item response.
	ItemsResponse struct {
		Data []Item `json:"data"`
	}

	// ItemsFilter is used to filter results from a List.
	ItemsFilter struct {
		// OwnerID filters for items owned by a given item.
		OwnerID *string

		// LocationID filters for items located in the given room.
		LocationID *string

		// InventoryID filters for items in the inventory of the given player.
		InventoryID *string

		// Restrict to a subset of the results.
		Offset int
		Limit  int
	}

	// ItemsStorage represents the persistent storage of items.
	ItemsStorage interface {
		// List returns a slice of items based on the value of the filter.
		List(ctx context.Context, filter ItemsFilter) ([]Item, error)

		// Get returns a single item given the itemID.
		Get(ctx context.Context, itemID string) (Item, error)

		// Create a item given the item request, returning the creating item.
		Create(ctx context.Context, req ItemRequest) (Item, error)

		// Update a item given the item request, returning the updated item.
		Update(ctx context.Context, itemID string, req ItemRequest) (Item, error)

		// Remove deletes the given item from persistent storage.
		Remove(ctx context.Context, itemID string) error
	}
)

// Validate returns an error for an invalid item request. A vaild request
// will return the parsed owner and location UUIDs.
func (r ItemRequest) Validate() (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	if r.Name == "" {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty item name", errors.ErrBadRequest)
	}
	if len(r.Name) > MaxItemNameLen {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: item name exceeds maximum length", errors.ErrBadRequest)
	}
	if r.Description == "" {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty item description", errors.ErrBadRequest)
	}
	if len(r.Description) > MaxItemDescriptionLen {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: item description exceeds maximum length", errors.ErrBadRequest)
	}
	ownerID, err := uuid.Parse(r.OwnerID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid ownerID: '%s'", errors.ErrBadRequest, r.OwnerID)
	}
	locationID, err := uuid.Parse(r.LocationID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid locationID: '%s'", errors.ErrBadRequest, r.LocationID)
	}
	inventoryID, err := uuid.Parse(r.InventoryID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid inventoryID: '%s'", errors.ErrBadRequest, r.InventoryID)
	}
	return ownerID, locationID, inventoryID, nil
}

// NewItemsResponse returns a items response given a slice of items.
func NewItemsResponse(is []Item) ItemsResponse {
	var resp ItemsResponse
	resp.Data = append(resp.Data, is...)
	return resp
}
