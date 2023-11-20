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

package rest // import "arcadium.dev/arcade/asset/rest"

import (
	"arcadium.dev/arcade/asset"
)

type (
	// ItemCreateRequest is used to request an item be created.
	//
	// swagger:parameters ItemCreate
	ItemCreateRequest struct {
		ItemRequest
	}

	// ItemUpdateRequest is used to request an item be updated.
	//
	// swagger:parameters ItemUpdate
	ItemUpdateRequest struct {
		ItemRequest
	}

	// ItemRequest is used to request an item be created or updated.
	ItemRequest struct {
		// Name is the name of the item.
		// in: body
		// minimum length: 1
		// maximum length: 256
		Name string `json:"name"`

		// Description is the description of the item.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		Description string `json:"description"`

		// OwnerID is the ID of the owner of the item.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		OwnerID string `json:"ownerID"`

		// LocationID is the ID of the location of the item.
		// in: body
		LocationID ItemLocationID `json:"locationID"`
	}

	// ItemResponse returns an item.
	ItemResponse struct {
		// Item returns the information about an item.
		// in: body
		Item Item `json:"item"`
	}

	// ItemsResponse returns multiple items.
	ItemsResponse struct {
		// Items returns the information about multiple items.
		// in: body
		Items []Item `json:"items"`
	}

	// Item holds an item's information, and is sent in a response.
	//
	// swagger:parameter
	Item struct {
		// ID is the item identifier.
		// in: body
		ID string `json:"id"`

		// Name is the item name.
		// in: body
		Name string `json:"name"`

		// Description is the item description.
		// in: body
		Description string `json:"description"`

		// OwnerID is the PlayerID of the item owner.
		// in:body
		OwnerID string `json:"ownerID"`

		// LocationID is the LocationID of the item's location.
		// in: body
		LocationID ItemLocationID `json:"locationID"`

		// Created is the time of the item's creation.
		// in: body
		Created asset.Timestamp `json:"created"`

		// Updated is the time the item was last updated.
		// in: body
		Updated asset.Timestamp `json:"updated"`
	}

	// ItemLocationID holds the locationID of the item, and the type of location.
	ItemLocationID struct {
		// ID is the location identifier. This can correspond the the ID of a room, player or item.
		// in: body
		ID string `json:"id"`

		// Type is the type of location. This can be "room", "player" or "item".
		Type string `json:"type"`
	}
)
