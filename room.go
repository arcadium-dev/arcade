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
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/core/errors"
)

const (
	MaxRoomNameLen          = 255
	MaxRoomDescriptionLen   = 4096
	DefaultRoomsFilterLimit = 10
	MaxRoomsFilterLimit     = 100
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
		OwnerID *uuid.UUID

		// ParentID filters for rooms located in a parent room (non-recursive).
		ParentID *uuid.UUID

		// Restrict to a subset of the results.
		Offset int
		Limit  int
	}

	// RoomsStorage represents the persistent storage of rooms.
	RoomsStorage interface {
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
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty room name", errors.ErrBadRequest)
	}
	if len(r.Name) > MaxRoomNameLen {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: room name exceeds maximum length", errors.ErrBadRequest)
	}
	if r.Description == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty room description", errors.ErrBadRequest)
	}
	if len(r.Description) > MaxRoomDescriptionLen {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: room description exceeds maximum length", errors.ErrBadRequest)
	}
	ownerID, err := uuid.Parse(r.OwnerID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid ownerID: '%s'", errors.ErrBadRequest, r.OwnerID)
	}
	parentID, err := uuid.Parse(r.ParentID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid parentID: '%s'", errors.ErrBadRequest, r.ParentID)
	}
	return ownerID, parentID, nil
}

// NewRoomsResponse returns a rooms response given a slice of rooms.
func NewRoomsResponse(rs []Room) RoomsResponse {
	var resp RoomsResponse
	resp.Data = append(resp.Data, rs...)
	return resp
}

// NewRoomsFilter creates a RoomsFilter from the the given request's URL
// query parameters
func NewRoomsFilter(r *http.Request) (RoomsFilter, error) {
	q := r.URL.Query()
	filter := RoomsFilter{
		Limit: DefaultRoomsFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return RoomsFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = &ownerID
	}

	if values := q["parentID"]; len(values) > 0 {
		parentID, err := uuid.Parse(values[0])
		if err != nil {
			return RoomsFilter{}, fmt.Errorf("%w: invalid parentID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.ParentID = &parentID
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > MaxRoomsFilterLimit {
			return RoomsFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = limit
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return RoomsFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = offset
	}

	return filter, nil
}
