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

package arcade // import "arcadium.dev/arcade"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/core/errors"
)

const (
	MaxRoomNameLen        = 255
	MaxRoomDescriptionLen = 4096
)

type (
	// Room is the internal representation of the data related to a room.
	Room struct {
		ID          string    `json:"roomID"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		OwnerID     string    `json:"ownerID"`
		ParentID    string    `json:"parentID"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
	}

	// RoomRequest is the payload of a room create or update request.
	RoomRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		OwnerID     string `json:"ownerID"`
		ParentID    string `json:"parentID"`
	}

	// RoomResponse is used to json encoded a single room response.
	RoomResponse struct {
		Data Room `json:"data"`
	}

	// RoomsResponse is used to json encoded a multi-room response.
	RoomsResponse struct {
		Data []Room `json:"data"`
	}

	// RoomsFilter is used to filter results from a List.
	RoomsFilter struct {
		// OwnerID filters for rooms owned by a given room.
		OwnerID *string

		// ParentID filters for rooms located in a parent room (non-recursive).
		ParentID *string

		// Restrict to a subset of the results.
		Offset int
		Limit  int
	}

	// RoomStorage represents the persistent storage of rooms.
	RoomStorage interface {
		// List returns a slice of rooms based on the value of the filter.
		List(ctx context.Context, filter RoomsFilter) ([]Room, error)

		// Get returns a single room given the roomID.
		Get(ctx context.Context, roomID string) (Room, error)

		// Create a room given the room request, returning the creating room.
		Create(ctx context.Context, req RoomRequest) (Room, error)

		// Update a room given the room request, returning the updated room.
		Update(ctx context.Context, roomID string, req RoomRequest) (Room, error)

		// Remove deletes the given room from persistent storage.
		Remove(ctx context.Context, roomID string) error
	}
)

// Validate returns an error for an invalid room request. A vaild request
// will return the parsed owner and parent UUIDs.
func (r RoomRequest) Validate() (uuid.UUID, uuid.UUID, error) {
	if r.Name == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty room name", errors.ErrInvalidArgument)
	}
	if len(r.Name) > MaxRoomNameLen {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: room name exceeds maximum length", errors.ErrInvalidArgument)
	}
	if r.Description == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty room description", errors.ErrInvalidArgument)
	}
	if len(r.Description) > MaxRoomDescriptionLen {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: room description exceeds maximum length", errors.ErrInvalidArgument)
	}
	ownerID, err := uuid.Parse(r.OwnerID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid ownerID: '%s'", errors.ErrInvalidArgument, r.OwnerID)
	}
	parentID, err := uuid.Parse(r.ParentID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid parentID: '%s'", errors.ErrInvalidArgument, r.ParentID)
	}
	return ownerID, parentID, nil
}

// NewRoomsResponse returns a rooms response given a slice of rooms.
func NewRoomsResponse(rs []Room) RoomsResponse {
	var resp RoomsResponse
	for _, r := range rs {
		resp.Data = append(resp.Data, r)
	}
	return resp
}
