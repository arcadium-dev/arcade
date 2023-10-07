//  Copyright 2022 arcadium.dev <info@arcadium.dev>
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

package assets // import "arcadium.dev/arcade/assets"

import (
	"github.com/google/uuid"
)

const (
	MaxLinkNameLen        = 256
	MaxLinkDescriptionLen = 4096

	DefaultLinksFilterLimit = 50
	MaxLinksFilterLimit     = 100
)

type (
	// LinkID is the unique identifier of a link.
	LinkID uuid.UUID
)

func (l LinkID) String() string { return uuid.UUID(l).String() }

type (
	// Link is the internal representation of a link.
	Link struct {
		ID            LinkID
		Name          string
		Description   string
		OwnerID       PlayerID
		LocationID    RoomID
		DestinationID RoomID
		Created       Timestamp
		Updated       Timestamp
	}

	// LinksFilter is used to filter results from a List.
	LinksFilter struct {
		// OwnerID filters for links owned by a given link.
		OwnerID PlayerID

		// LocationID filters for links located in a location link.
		LocationID RoomID

		// DestinationID filters for links connected to the given destination.
		DestinationID RoomID

		// Offset is used to restrict to a subset of the results,
		// indicating the initial offset into the set of results.
		Offset uint

		// Limit is used to restrict to a subset of the results,
		// indicating the maximum number of results to return.
		Limit uint
	}

	// LinkCreateRequest is used to request an item to be created.
	LinkCreateRequest = LinkRequest

	// LinkCreateRequest is used to request an item to be updated.
	LinkUpdateRequest = LinkRequest

	// LinkRequest holds item information.
	LinkRequest struct {
		Name          string
		Description   string
		OwnerID       PlayerID
		LocationID    RoomID
		DestinationID RoomID
	}
)
