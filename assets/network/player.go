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

package network // import "arcadium.dev/arcade/assets/network"

import (
	"arcadium.dev/arcade/assets"
)

type (
	// PlayerCreateRequest is used to request an player be created.
	//
	// swagger:parameters PlayerCreate
	PlayerCreateRequest = PlayerRequest

	// PlayerUpdateRequest is used to request an player be updated.
	//
	// swagger:parameters PlayerUpdate
	PlayerUpdateRequest = PlayerRequest

	// PlayerRequest is used to request an player be created or updated.
	PlayerRequest struct {
		// Name is the name of the player.
		// in: body
		// minimum length: 1
		// maximum length: 256
		Name string `json:"name"`

		// Description is the description of the player.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		Description string `json:"description"`

		// HomeID is the ID of the home of the player.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		HomeID string `json:"homeID"`

		// LocationID is the ID of the location of the player.
		// in: body
		LocationID string `json:"locationID"`
	}

	// PlayerResponse returns a player.
	PlayerResponse struct {
		// Player returns the information about a player.
		// in: body
		Player Player `json:"player"`
	}

	// PlayersResponse returns multiple players.
	PlayersResponse struct {
		// Players returns the information about multiple players.
		// in: body
		Players []Player `json:"players"`
	}

	// Player holds a player's information, and is sent in a response.
	//
	// swagger:parameter
	Player struct {
		// ID is the player identifier.
		// in: body
		ID string `json:"id"`

		// Name is the player name.
		// in: body
		Name string `json:"name"`

		// Description is the player description.
		// in: body
		Description string `json:"description"`

		// HomeID is the RoomID of the player's home.
		// in:body
		HomeID string `json:"homeID"`

		// LocationID is the RoomID of the player's location.
		// in: body
		LocationID string `json:"locationID"`

		// Created is the time of the player's creation.
		// in: body
		Created assets.Timestamp `json:"created"`

		// Updated is the time the player was last updated.
		// in: body
		Updated assets.Timestamp `json:"updated"`
	}
)
