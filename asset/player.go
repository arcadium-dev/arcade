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

package asset // import "arcadium.dev/arcade/asset"

import (
	"database/sql/driver"

	"github.com/google/uuid"
)

const (
	MaxPlayerNameLen        = 256
	MaxPlayerDescriptionLen = 4096

	DefaultPlayerFilterLimit = 50
	MaxPlayerFilterLimit     = 100
)

type (
	// PlayerID is the unique identifier of an player.
	PlayerID uuid.UUID
)

var (
	NilPlayerID = PlayerID(uuid.Nil)
)

func (p PlayerID) ID() LocationID               { return LocationID(p) }
func (p PlayerID) Type() LocationType           { return LocationTypePlayer }
func (p PlayerID) String() string               { return uuid.UUID(p).String() }
func (p *PlayerID) Scan(src any) error          { return (*uuid.UUID)(p).Scan(src) }
func (p PlayerID) Value() (driver.Value, error) { return uuid.UUID(p).Value() }

type (
	// Player is the internal representation of the data related to a player.
	Player struct {
		ID          PlayerID
		Name        string
		Description string
		HomeID      RoomID
		LocationID  RoomID
		Created     Timestamp
		Updated     Timestamp
	}

	// PlayerFilter is used to filter results from List.
	PlayerFilter struct {
		// LocationID filters for players in a given location.
		LocationID RoomID

		// Offset is used to restrict to a subset of the results,
		// indicating the initial offset into the set of results.
		Offset uint

		// Limit is used to restrict to a subset of the results,
		// indicating the maximum number of results to return.
		Limit uint
	}

	// PlayerCreate is used to create an item.
	PlayerCreate struct {
		PlayerChange
	}

	// PlayerUpdate is used to update an item.
	PlayerUpdate struct {
		PlayerChange
	}

	// PlayerChange holds information to change an item.
	PlayerChange struct {
		Name        string
		Description string
		HomeID      RoomID
		LocationID  RoomID
	}
)
