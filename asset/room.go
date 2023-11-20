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
	MaxRoomNameLen        = 256
	MaxRoomDescriptionLen = 4096

	DefaultRoomFilterLimit = 50
	MaxRoomFilterLimit     = 100
)

type (
	// RoomID is the unique identifier of an room.
	RoomID uuid.UUID
)

var (
	NilRoomID = RoomID(uuid.Nil)
)

func (r RoomID) ID() LocationID               { return LocationID(r) }
func (r RoomID) Type() LocationType           { return LocationTypeRoom }
func (r RoomID) String() string               { return uuid.UUID(r).String() }
func (r *RoomID) Scan(src any) error          { return (*uuid.UUID)(r).Scan(src) }
func (r RoomID) Value() (driver.Value, error) { return uuid.UUID(r).Value() }

type (
	// Room is the internal representation of the data related to a room.
	Room struct {
		ID          RoomID
		Name        string
		Description string
		OwnerID     PlayerID
		ParentID    RoomID
		Created     Timestamp
		Updated     Timestamp
	}

	// RoomFilter is used to filter results from a List.
	RoomFilter struct {
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

	// RoomCreate is used to create an item.
	RoomCreate struct {
		RoomChange
	}

	// RoomUpdate is used to update an item.
	RoomUpdate struct {
		RoomChange
	}

	// RoomChange holds information to change an item.
	RoomChange struct {
		Name        string
		Description string
		OwnerID     PlayerID
		ParentID    RoomID
	}
)
