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

package players

import (
	"time"

	"arcadium.dev/arcade/internal/arcade"
)

const (
	maxNameLen        = 255
	maxDescriptionLen = 4096
)

type (
	// player is the internal representation of the data related to a player.
	player struct {
		id          string
		name        string
		description string
		home        string
		location    string
		created     time.Time
		updated     time.Time
	}

	// playerRequest is the payload of a player request.
	playerRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Home        string `json:"home"`
		Location    string `json:"location"`
	}

	// playerResponse is used as payload data for player responses.
	playerResponseData struct {
		PlayerID    string    `json:"playerID"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Home        string    `json:"home"`
		Location    string    `json:"location"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
	}

	// playerResponse is used to json encoded a response with a single player.
	playerResponse struct {
		Data playerResponseData `json:"data"`
	}

	// playersResponse is used to json encoded a response with a multiple players.
	playersResponse struct {
		Data []playerResponseData `json:"data"`
	}
)

func newPlayer(p playerRequest) arcade.Player {
	return player{
		name:        p.Name,
		description: p.Description,
		home:        p.Home,
		location:    p.Location,
	}
}

func (p player) ID() string          { return p.id }
func (p player) Name() string        { return p.name }
func (p player) Description() string { return p.description }
func (p player) Home() string        { return p.home }
func (p player) Location() string    { return p.location }
func (p player) Created() time.Time  { return p.created }
func (p player) Updated() time.Time  { return p.updated }

func newPlayerResponseData(p arcade.Player) playerResponseData {
	return playerResponseData{
		PlayerID:    p.ID(),
		Name:        p.Name(),
		Description: p.Description(),
		Home:        p.Home(),
		Location:    p.Location(),
		Created:     p.Created(),
		Updated:     p.Updated(),
	}
}

func newPlayerResponse(p arcade.Player) playerResponse {
	return playerResponse{
		Data: newPlayerResponseData(p),
	}
}

func newPlayersResponse(ps []arcade.Player) playersResponse {
	var r playersResponse
	for _, p := range ps {
		r.Data = append(r.Data, newPlayerResponseData(p))
	}
	return r
}
