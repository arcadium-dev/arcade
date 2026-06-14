//  Copyright 2022-2026 arcadium.dev <info@arcadium.dev>
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
	"arcadium.dev/arcade"
)

type (
	// LinkCreateRequest is used to request an link be created.
	LinkCreateRequest struct {
		LinkRequest
	}

	// LinkUpdateRequest is used to request an link be updated.
	LinkUpdateRequest struct {
		LinkRequest
	}

	// LinkRequest is used to request an link be created or updated.
	LinkRequest struct {
		// Name is the name of the link.
		Name string `json:"name"`

		// Description is the description of the link.
		Description string `json:"description"`

		// OwnerID is the ID of the owner of the link.
		OwnerID string `json:"ownerID"`

		// LocationID is the ID of the location of the link.
		LocationID string `json:"locationID"`

		// DestinationID is the ID of the destination of the link.
		DestinationID string `json:"destinationID"`
	}

	// LinkResponse returns a link.
	LinkResponse struct {
		// Link returns the information about a link.
		Link Link `json:"link"`
	}

	// LinksResponse returns multiple links.
	LinksResponse struct {
		// Links returns the information about multiple links.
		Links []Link `json:"links"`
	}

	// Link holds a link's information, and is sent in a response.
	Link struct {
		// ID is the link identifier.
		ID string `json:"id"`

		// Name is the link name.
		Name string `json:"name"`

		// Description is the link description.
		Description string `json:"description"`

		// OwnerID is the PlayerID of the link owner.
		OwnerID string `json:"ownerID"`

		// LocationID is the RoomID of the link's location.
		LocationID string `json:"locationID"`

		// DestinationID is the RoomID of the link's destination.
		DestinationID string `json:"destinationID"`

		// Created is the time of the link's creation.
		Created arcade.Timestamp `json:"created"`

		// Updated is the time the link was last updated.
		Updated arcade.Timestamp `json:"updated"`
	}
)
