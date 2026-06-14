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
	// PlayerCreateRequest is used to request an player be created.
	PlayerCreateRequest struct {
		PlayerRequest
	}

	// PlayerUpdateRequest is used to request an player be updated.
	PlayerUpdateRequest struct {
		PlayerRequest
	}

	// PlayerRequest is used to request an player be created or updated.
	PlayerRequest struct {
		// Name is the name of the player.
		Name string `json:"name"`

		// Description is the description of the player.
		Description string `json:"description"`

		// HomeID is the ID of the home of the player.
		HomeID string `json:"homeID"`

		// LocationID is the ID of the location of the player.
		LocationID string `json:"locationID"`
	}

	// PlayerResponse returns a player.
	PlayerResponse struct {
		// Player returns the information about a player.
		Player Player `json:"player"`
	}

	// PlayersResponse returns multiple players.
	PlayersResponse struct {
		// Players returns the information about multiple players.
		Players []Player `json:"players"`
	}

	// Player holds a player's information, and is sent in a response.
	Player struct {
		// ID is the player identifier.
		ID string `json:"id"`

		// Name is the player name.
		Name string `json:"name"`

		// Description is the player description.
		Description string `json:"description"`

		// HomeID is the RoomID of the player's home.
		HomeID string `json:"homeID"`

		// LocationID is the RoomID of the player's location.
		LocationID string `json:"locationID"`

		// Created is the time of the player's creation.
		Created arcade.Timestamp `json:"created"`

		// Updated is the time the player was last updated.
		Updated arcade.Timestamp `json:"updated"`
	}
)
