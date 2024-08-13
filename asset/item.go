//  Copyright 2022-2024 arcadium.dev <info@arcadium.dev>
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

	"arcadium.dev/arcade"
)

const (
	MaxItemNameLen        = 256
	MaxItemDescriptionLen = 4096

	DefaultItemFilterLimit = 50
	MaxItemFilterLimit     = 100
)

type (
	// ItemID is the unique identifier of an item.
	ItemID uuid.UUID
)

func (i ItemID) ID() LocationID               { return LocationID(i) }
func (i ItemID) Type() LocationType           { return LocationTypeItem }
func (i ItemID) String() string               { return uuid.UUID(i).String() }
func (i *ItemID) Scan(src any) error          { return (*uuid.UUID)(i).Scan(src) }
func (i ItemID) Value() (driver.Value, error) { return uuid.UUID(i).Value() }

type (
	// Item is the internal representation of an item.
	Item struct {
		ID          ItemID
		Name        string
		Description string
		OwnerID     PlayerID
		LocationID  ItemLocationID
		Created     arcade.Timestamp
		Updated     arcade.Timestamp
	}

	// ItemFilter is used to filter results from a list of all items.
	ItemFilter struct {
		// OwnerID filters for items owned by the given player.
		OwnerID PlayerID

		// LocationID filters for items in the given location.
		LocationID ItemLocationID

		// Offset is used to restrict to a subset of the results,
		// indicating the initial offset into the set of results.
		Offset uint

		// Limit is used to restrict to a subset of the results,
		// indicating the maximum number of results to return.
		Limit uint
	}

	// ItemCreate is used to create an item.
	ItemCreate struct {
		ItemChange
	}

	// ItemUpdate is used to update an item.
	ItemUpdate struct {
		ItemChange
	}

	// ItemChange holds information to change an item.
	ItemChange struct {
		Name        string
		Description string
		OwnerID     PlayerID
		LocationID  ItemLocationID
	}
)

type (
	// ItemLocationID defines the expected behavior of an object that contain an item.
	ItemLocationID interface {
		ID() LocationID
		Type() LocationType
	}

	// LocationID provides the ID of the item's location.
	LocationID uuid.UUID

	// LocationType provides the type of location, room, player, or item.
	LocationType uint8
)

var (
	NilLocationID = LocationID(uuid.Nil)
)

func (l LocationID) String() string               { return uuid.UUID(l).String() }
func (l *LocationID) Scan(src any) error          { return (*uuid.UUID)(l).Scan(src) }
func (l LocationID) Value() (driver.Value, error) { return uuid.UUID(l).Value() }

const (
	LocationTypeRoom = LocationType(iota)
	LocationTypePlayer
	LocationTypeItem
)

func (t LocationType) String() string {
	switch t {
	case LocationTypeRoom:
		return "room"
	case LocationTypePlayer:
		return "player"
	}
	return "item"
}
