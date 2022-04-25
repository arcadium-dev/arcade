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

package rooms

import (
	"time"

	"arcadium.dev/arcade/internal/arcade"
)

const (
	maxNameLen        = 255
	maxDescriptionLen = 4096
)

type (
	// room is the internal representation of the data related to a room.
	room struct {
		id          string
		name        string
		description string
		owner       string
		parent      string
		created     time.Time
		updated     time.Time
	}

	// roomRequest is the payload of a room request.
	roomRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       string `json:"owner"`
		Parent      string `json:"parent"`
	}

	// roomResponse is used as payload data for room responses.
	roomResponseData struct {
		RoomID      string    `json:"roomID"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Owner       string    `json:"owner"`
		Parent      string    `json:"parent"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
	}

	// roomResponse is used to json encoded a response with a single room.
	roomResponse struct {
		Data roomResponseData `json:"data"`
	}

	// roomsResponse is used to json encoded a response with a multiple rooms.
	roomsResponse struct {
		Data []roomResponseData `json:"data"`
	}
)

func newRoom(p roomRequest) arcade.Room {
	return room{
		name:        p.Name,
		description: p.Description,
		owner:       p.Owner,
		parent:      p.Parent,
	}
}

func (p room) ID() string          { return p.id }
func (p room) Name() string        { return p.name }
func (p room) Description() string { return p.description }
func (p room) Owner() string       { return p.owner }
func (p room) Parent() string      { return p.parent }
func (p room) Created() time.Time  { return p.created }
func (p room) Updated() time.Time  { return p.updated }

func newRoomResponseData(p arcade.Room) roomResponseData {
	return roomResponseData{
		RoomID:      p.ID(),
		Name:        p.Name(),
		Description: p.Description(),
		Owner:       p.Owner(),
		Parent:      p.Parent(),
		Created:     p.Created(),
		Updated:     p.Updated(),
	}
}

func newRoomResponse(p arcade.Room) roomResponse {
	return roomResponse{
		Data: newRoomResponseData(p),
	}
}

func newRoomsResponse(ps []arcade.Room) roomsResponse {
	var r roomsResponse
	for _, p := range ps {
		r.Data = append(r.Data, newRoomResponseData(p))
	}
	return r
}
