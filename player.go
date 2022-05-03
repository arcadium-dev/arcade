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
	MaxPlayerNameLen        = 255
	MaxPlayerDescriptionLen = 4096
)

type (
	// Player is the internal representation of the data related to a player.
	Player struct {
		ID          string    `json:"playerID"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		HomeID      string    `json:"homeID"`
		LocationID  string    `json:"locationID"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
	}

	// PlayerRequest is the payload of a player create or update request.
	PlayerRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		HomeID      string `json:"homeID"`
		LocationID  string `json:"locationID"`
	}

	// PlayerResponse is used to json encoded a single player response.
	PlayerResponse struct {
		Data Player `json:"data"`
	}

	// PlayersResponse is used to json encoded a multi-player resposne.
	PlayersResponse struct {
		Data []Player `json:"data"`
	}

	// PlayersFilter is used to filter results from List.
	PlayersFilter struct {
		// LocationID filters for players in a given location.
		LocationID string

		// Restrict to a subset of the results.
		Offset int
		Limit  int
	}

	// PlayerStorage represents the persistent storage of players.
	PlayerStorage interface {
		// List returns a slice of players based on the value of the filter.
		List(ctx context.Context, filter PlayersFilter) ([]Player, error)

		// Get returns a single player given the playerID.
		Get(ctx context.Context, playerID string) (Player, error)

		// Create a player given the player request, returning the creating player.
		Create(ctx context.Context, req PlayerRequest) (Player, error)

		// Update a player given the player request, returning the updated player.
		Update(ctx context.Context, playerID string, req PlayerRequest) (Player, error)

		// Remove deletes the given player from persistent storage.
		Remove(ctx context.Context, playerID string) error
	}
)

// Validate returns an error for an invalid player request. A vaild request
// will return the parsed home and location UUIDs.
func (r PlayerRequest) Validate() (uuid.UUID, uuid.UUID, error) {
	if r.Name == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty player name", errors.ErrInvalidArgument)
	}
	if len(r.Name) > MaxPlayerNameLen {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: player name exceeds maximum length", errors.ErrInvalidArgument)
	}
	if r.Description == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty player description", errors.ErrInvalidArgument)
	}
	if len(r.Description) > MaxPlayerDescriptionLen {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: player description exceeds maximum length", errors.ErrInvalidArgument)
	}
	homeID, err := uuid.Parse(r.HomeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid homeID: '%s'", errors.ErrInvalidArgument, r.HomeID)
	}
	locationID, err := uuid.Parse(r.LocationID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid locationID: '%s'", errors.ErrInvalidArgument, r.LocationID)
	}
	return homeID, locationID, nil
}

// NewPlayersResponse returns a players response given a slice of players.
func NewPlayersResponse(ps []Player) PlayersResponse {
	var resp PlayersResponse
	for _, p := range ps {
		resp.Data = append(resp.Data, p)
	}
	return resp
}
