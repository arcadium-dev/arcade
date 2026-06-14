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
	// RoomCreateRequest is used to request an room be created.
	RoomCreateRequest struct {
		RoomRequest
	}

	// RoomUpdateRequest is used to request an room be updated.
	RoomUpdateRequest struct {
		RoomRequest
	}

	// RoomRequest is used to request an item be created or updated.
	RoomRequest struct {
		// Name is the name of the room.
		Name string `json:"name"`

		// Description is the description of the room.
		Description string `json:"description"`

		// OwnerID is the ID of the owner of the room.
		OwnerID string `json:"ownerID"`

		// ParentID is the ID of the parent of the room.
		ParentID string `json:"parentID"`
	}

	// RoomResponse returns a room.
	RoomResponse struct {
		// Room returns the information about a room.
		Room Room `json:"room"`
	}

	// RoomsResponse returns multiple rooms.
	RoomsResponse struct {
		// Rooms returns the information about multiple rooms.
		Rooms []Room `json:"rooms"`
	}

	// Room holds a room's information, and is sent in a response.
	Room struct {
		// ID is the room identifier.
		ID string `json:"id"`

		// Name is the room name.
		Name string `json:"name"`

		// Description is the room description.
		Description string `json:"description"`

		// OwnerID is the PlayerID of the room owner.
		OwnerID string `json:"ownerID"`

		// ParentID is the RoomID of the room's parent.
		ParentID string `json:"parentID"`

		// Created is the time of the room's creation.
		Created arcade.Timestamp `json:"created"`

		// Updated is the time the room was last updated.
		Updated arcade.Timestamp `json:"updated"`
	}
)
