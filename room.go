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
	"time"

	"github.com/google/uuid"
)

const (
	MaxRoomNameLen        = 256
	MaxRoomDescriptionLen = 4096

	DefaultRoomsFilterLimit = 50
	MaxRoomsFilterLimit     = 100
)

type (
	// RoomID is the unique identifier of an room.
	RoomID uuid.UUID
)

func (r RoomID) ID() LocationID     { return LocationID(r) }
func (r RoomID) Type() LocationType { return LocationTypeRoom }
func (r RoomID) String() string     { return uuid.UUID(r).String() }

type (
	// Room is the internal representation of the data related to a room.
	Room struct {
		RoomID      RoomID
		Name        string
		Description string
		OwnerID     PlayerID
		ParentID    RoomID
		Created     time.Time
		Updated     time.Time
	}

	// RoomsFilter is used to filter results from a List.
	RoomsFilter struct {
		// OwnerID filters for rooms owned by a given room.
		OwnerID PlayerID

		// ParentID filters for rooms located in a parent room (non-recursive).
		ParentID RoomID

		// Offset is used to restrict to a subset of the results,
		// indicating the initial offset into the set of results.
		Offset uint

		// Limit is used to restrict to a subset of the results,
		// indicating the maximum number of results to return.
		Limit uint
	}
)
