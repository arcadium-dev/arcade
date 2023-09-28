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
	"github.com/google/uuid"
)

const (
	MaxPlayerNameLen        = 255
	MaxPlayerDescriptionLen = 4096

	DefaultPlayersFilterLimit = 10
	MaxPlayersFilterLimit     = 100
)

type (
	// PlayerID is the unique identifier of an player.
	PlayerID uuid.NullUUID
)

func (p PlayerID) ID() LocationID     { return LocationID(p) }
func (p PlayerID) Type() LocationType { return LocationTypePlayer }

type (
	// Player is the internal representation of the data related to a player.
	Player struct {
		PlayerID    PlayerID
		Name        string
		Description string
		HomeID      RoomID
		LocationID  RoomID
		Created     Timestamp
		Updated     Timestamp
	}

	// PlayersFilter is used to filter results from List.
	PlayersFilter struct {
		// LocationID filters for players in a given location.
		LocationID RoomID

		// Offset is used to restrict to a subset of the results,
		// indicating the initial offset into the set of results.
		Offset uint

		// Limit is used to restrict to a subset of the results,
		// indicating the maximum number of results to return.
		Limit uint
	}
)
