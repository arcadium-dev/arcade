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

package server // import "arcadium.dev/arcade/asset/server"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
)

const (
	V1PlayersRoute string = "/v1/players"
)

type (
	// PlayerService services player related network requests.
	PlayersService struct {
		Storage PlayerStorage
	}

	// PlayerStorage defines the expected behavior of the player manager in the domain layer.
	PlayerStorage interface {
		List(context.Context, asset.PlayerFilter) ([]*asset.Player, error)
		Get(context.Context, asset.PlayerID) (*asset.Player, error)
		Create(context.Context, asset.PlayerCreate) (*asset.Player, error)
		Update(context.Context, asset.PlayerID, asset.PlayerUpdate) (*asset.Player, error)
		Remove(context.Context, asset.PlayerID) error
	}
)

// Register sets up the http handler for this service with the given router.
func (s PlayersService) Register(router *mux.Router) {
	r := router.PathPrefix(V1PlayersRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{id}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{id}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{id}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (PlayersService) Name() string {
	return "players"
}

// Shutdown is a no-op since there no long running processes for this service.
func (PlayersService) Shutdown() {}

// List handles a request to retrieve multiple players.
func (s PlayersService) List(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/players PlayerList
	//
	// List returns a list of players.
	//
	// Produces: application/json
	//
	// Parameters:
	//   + name: locationID
	//     in: query
	//   + name: offset
	//     in: query
	//   + name: limit
	//     in: query
	//
	// Responses:
	//  200: PlayerResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Create a filter from the quesry parameters.
	filter, err := NewPlayerFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of players.
	aPlayers, err := s.Storage.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from asset players, to network players.
	players := make([]rest.Player, 0)
	for _, aPlayer := range aPlayers {
		players = append(players, TranslatePlayer(aPlayer))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.PlayersResponse{Players: players})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve a player.
func (s PlayersService) Get(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/players/{playerID} PlayerGet
	//
	// Get returns a player.
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: player ID
	//     required: true
	//
	// Responses:
	//  200: PlayerResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the playerID from the uri.
	id := mux.Vars(r)["id"]
	playerID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid player id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Request the player from the player manager.
	player, err := s.Storage.Get(ctx, asset.PlayerID(playerID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the player to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.PlayerResponse{Player: TranslatePlayer(player)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create a player.
func (s PlayersService) Create(w http.ResponseWriter, r *http.Request) {
	// swagger:route POST /v1/players PlayerCreate
	//
	// Create will create a new player based on the player request in the body of the
	// request.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Responses:
	//  200: PlayerResponse
	//  400: ResponseError
	//  409: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the player request from the body of the request.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request body: %s", errors.ErrBadRequest, err,
		))
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid json: a json encoded body is required", errors.ErrBadRequest,
		))
		return
	}

	var createReq rest.PlayerCreateRequest
	err = json.Unmarshal(body, &createReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the player request to the player manager.
	change, err := TranslatePlayerRequest(createReq.PlayerRequest)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	player, err := s.Storage.Create(ctx, asset.PlayerCreate{PlayerChange: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned player for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.PlayerResponse{Player: TranslatePlayer(player)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update a player.
func (s PlayersService) Update(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /v1/players/{id} PlayerUpdate
	//
	// Update will update player based on the playerID and the player\ request in the
	// body of the request.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: player ID
	//     required: true
	//
	// Responses:
	//  200: PlayerResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Grab the playerID from the uri.
	id := mux.Vars(r)["id"]
	playerID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid player id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Process the request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request body: %s", errors.ErrBadRequest, err,
		))
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid json: a json encoded body is required", errors.ErrBadRequest,
		))
		return
	}

	// Populate the network player from the body.
	var updateReq rest.PlayerUpdateRequest
	err = json.Unmarshal(body, &updateReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the player request.
	change, err := TranslatePlayerRequest(updateReq.PlayerRequest)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the player to the player manager.
	player, err := s.Storage.Update(ctx, asset.PlayerID(playerID), asset.PlayerUpdate{PlayerChange: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.PlayerResponse{Player: TranslatePlayer(player)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove a player.
func (s PlayersService) Remove(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /v1/players/{id} PlayerRemove
	//
	// Remove deletes the player.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: player ID
	//     required: true
	//
	// Responses:
	//  200: PlayerResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the playerID from the uri.
	id := mux.Vars(r)["id"]
	playerID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid player id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Send the playerID to the player manager for removal.
	err = s.Storage.Remove(ctx, asset.PlayerID(playerID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewPlayerFilter creates an asset players filter from the given request's URL query parameters.
func NewPlayerFilter(r *http.Request) (asset.PlayerFilter, error) {
	q := r.URL.Query()
	filter := asset.PlayerFilter{
		Limit: asset.DefaultPlayerFilterLimit,
	}

	if values := q["locationID"]; len(values) > 0 {
		locationID, err := uuid.Parse(values[0])
		if err != nil {
			return asset.PlayerFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.LocationID = asset.RoomID(locationID)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > asset.MaxPlayerFilterLimit {
			return asset.PlayerFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return asset.PlayerFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	return filter, nil
}

// TranslatesPlayerRequest translates a network request to an asset player request.
func TranslatePlayerRequest(p rest.PlayerRequest) (asset.PlayerChange, error) {
	empty := asset.PlayerChange{}

	if p.Name == "" {
		return empty, fmt.Errorf("%w: empty player name", errors.ErrBadRequest)
	}
	if len(p.Name) > asset.MaxPlayerNameLen {
		return empty, fmt.Errorf("%w: player name exceeds maximum length", errors.ErrBadRequest)
	}
	if p.Description == "" {
		return empty, fmt.Errorf("%w: empty player description", errors.ErrBadRequest)
	}
	if len(p.Description) > asset.MaxPlayerDescriptionLen {
		return empty, fmt.Errorf("%w: player description exceeds maximum length", errors.ErrBadRequest)
	}
	homeID, err := uuid.Parse(p.HomeID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid homeID: '%s'", errors.ErrBadRequest, p.HomeID)
	}
	locID, err := uuid.Parse(p.LocationID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid locationID: '%s', %s", errors.ErrBadRequest, p.LocationID, err)
	}

	return asset.PlayerChange{
		Name:        p.Name,
		Description: p.Description,
		HomeID:      asset.RoomID(homeID),
		LocationID:  asset.RoomID(locID),
	}, nil
}

// TranslatePlayer translates an arcade player to a network player.
func TranslatePlayer(p *asset.Player) rest.Player {
	return rest.Player{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		HomeID:      p.HomeID.String(),
		LocationID:  p.LocationID.String(),
		Created:     p.Created,
		Updated:     p.Updated,
	}
}
