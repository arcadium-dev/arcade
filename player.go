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
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/core/errors"
)

const (
	MaxPlayerNameLen          = 255
	MaxPlayerDescriptionLen   = 4096
	DefaultPlayersFilterLimit = 10
	MaxPlayersFilterLimit     = 100
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
		LocationID *uuid.UUID

		// Restrict to a subset of the results.
		Offset int
		Limit  int
	}

	// PlayersStorage represents the persistent storage of players.
	PlayersStorage interface {
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
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty player name", errors.ErrBadRequest)
	}
	if len(r.Name) > MaxPlayerNameLen {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: player name exceeds maximum length", errors.ErrBadRequest)
	}
	if r.Description == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty player description", errors.ErrBadRequest)
	}
	if len(r.Description) > MaxPlayerDescriptionLen {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: player description exceeds maximum length", errors.ErrBadRequest)
	}
	homeID, err := uuid.Parse(r.HomeID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid homeID: '%s'", errors.ErrBadRequest, r.HomeID)
	}
	locationID, err := uuid.Parse(r.LocationID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid locationID: '%s'", errors.ErrBadRequest, r.LocationID)
	}
	return homeID, locationID, nil
}

// NewPlayersResponse returns a players response given a slice of players.
func NewPlayersResponse(ps []Player) PlayersResponse {
	var resp PlayersResponse
	resp.Data = append(resp.Data, ps...)
	return resp
}

// NewPlayersFilter creates a PlayersFilter from the the given request's URL
// query parameters
func NewPlayersFilter(r *http.Request) (PlayersFilter, error) {
	q := r.URL.Query()
	filter := PlayersFilter{
		Limit: DefaultPlayersFilterLimit,
	}

	if values := q["locationID"]; len(values) > 0 {
		locationID, err := uuid.Parse(values[0])
		if err != nil {
			return PlayersFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.LocationID = &locationID
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > MaxPlayersFilterLimit {
			return PlayersFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = limit
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return PlayersFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = offset
	}

	return filter, nil
}
