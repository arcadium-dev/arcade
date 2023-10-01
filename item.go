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
	MaxItemNameLen        = 256
	MaxItemDescriptionLen = 4096

	DefaultItemsFilterLimit = 50
	MaxItemsFilterLimit     = 100
)

type (
	// ItemID is the unique identifier of an item.
	ItemID uuid.UUID
)

func (i ItemID) ID() LocationID     { return LocationID(i) }
func (i ItemID) Type() LocationType { return LocationTypeItem }
func (i ItemID) String() string     { return uuid.UUID(i).String() }

type (
	// Item is the internal representation of an item.
	Item struct {
		ID          ItemID
		Name        string
		Description string
		OwnerID     PlayerID
		LocationID  ItemLocationID
		Created     Timestamp
		Updated     Timestamp
	}

	// ItemsFilter is used to filter results from a list of all items.
	ItemsFilter struct {
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

	// IngressItem is used to request an item be created or updated.
	IngressItem struct {
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

	// LocationType provides the type of location, room, player, or item.
	LocationType uint8

	// LocationID provides the ID of the item's location.
	LocationID uuid.UUID
)

func (l LocationID) String() string { return uuid.UUID(l).String() }

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
