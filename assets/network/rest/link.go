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

package rest // import "arcadium.dev/arcade/assets/network/rest"

import "arcadium.dev/arcade/assets"

type (
	// LinkCreateRequest is used to request an link be created.
	//
	// swagger:parameters LinkCreate
	LinkCreateRequest struct {
		LinkRequest
	}

	// LinkUpdateRequest is used to request an link be updated.
	//
	// swagger:parameters LinkUpdate
	LinkUpdateRequest struct {
		LinkRequest
	}

	// LinkRequest is used to request an link be created or updated.
	LinkRequest struct {
		// Name is the name of the link.
		// in: body
		// minimum length: 1
		// maximum length: 256
		Name string `json:"name"`

		// Description is the description of the link.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		Description string `json:"description"`

		// OwnerID is the ID of the owner of the link.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		OwnerID string `json:"ownerID"`

		// LocationID is the ID of the location of the link.
		// in: body
		LocationID string `json:"locationID"`

		// DestinationID is the ID of the destination of the link.
		// in: body
		DestinationID string `json:"destinationID"`
	}

	// LinkResponse returns a link.
	LinkResponse struct {
		// Link returns the information about a link.
		// in: body
		Link Link `json:"link"`
	}

	// LinksResponse returns multiple links.
	LinksResponse struct {
		// Links returns the information about multiple links.
		// in: body
		Links []Link `json:"links"`
	}

	// Link holds a link's information, and is sent in a response.
	//
	// swagger:parameter
	Link struct {
		// ID is the link identifier.
		// in: body
		ID string `json:"id"`

		// Name is the link name.
		// in: body
		Name string `json:"name"`

		// Description is the link description.
		// in: body
		Description string `json:"description"`

		// OwnerID is the PlayerID of the link owner.
		// in:body
		OwnerID string `json:"ownerID"`

		// LocationID is the RoomID of the link's location.
		// in: body
		LocationID string `json:"locationID"`

		// DestinationID is the RoomID of the link's destination.
		// in: body
		DestinationID string `json:"destinationID"`

		// Created is the time of the link's creation.
		// in: body
		Created assets.Timestamp `json:"created"`

		// Updated is the time the link was last updated.
		// in: body
		Updated assets.Timestamp `json:"updated"`
	}
)
