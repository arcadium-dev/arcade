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

package networking // import "arcadium.dev/networking"

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

	"arcadium.dev/arcade"
)

const (
	V1PlayersRoute string = "/v1/players"
)

type (
	// PlayerService services player related network requests.
	PlayersService struct {
		Manager PlayerManager
	}

	// PlayerManager defines the expected behavior of the player manager in the domain layer.
	PlayerManager interface {
		List(ctx context.Context, filter arcade.PlayersFilter) ([]*arcade.Player, error)
		Get(ctx context.Context, playerID arcade.PlayerID) (*arcade.Player, error)
		Create(ctx context.Context, ingressPlayer arcade.IngressPlayer) (*arcade.Player, error)
		Update(ctx context.Context, playerID arcade.PlayerID, ingressPlayer arcade.IngressPlayer) (*arcade.Player, error)
		Remove(ctx context.Context, playerID arcade.PlayerID) error
	}

	// IngressPlayer is used to request a player be created or updated.
	//
	// swagger:parameters PlayerCreate PlayerUpdate
	IngressPlayer struct {
		// Name is the name of the player.
		// in: body
		// minimum length: 1
		// maximum length: 256
		Name string `json:"name"`

		// Description is the description of the player.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		Description string `json:"description"`

		// HomeID is the ID of the home of the player.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		HomeID string `json:"ownerID"`

		// LocationID is the ID of the location of the player.
		// in: body
		LocationID string `json:"locationID"`
	}

	// EgressPlayer returns a player.
	EgressPlayer struct {
		// Player returns the information about a player.
		// in: body
		Player Player `json:"player"`
	}

	// PlayersResponse returns multiple players.
	EgressPlayers struct {
		// Players returns the information about multiple players.
		// in: body
		Players []Player `json:"players"`
	}

	// Player holds a player's information, and is sent in a response.
	//
	// swagger:parameter
	Player struct {
		// ID is the player identifier.
		// in: body
		ID string `json:"id"`

		// Name is the player name.
		// in: body
		Name string `json:"name"`

		// Description is the player description.
		// in: body
		Description string `json:"description"`

		// HomeID is the RoomID of the player's home.
		// in:body
		HomeID string `json:"ownerID"`

		// LocationID is the RoomID of the player's location.
		// in: body
		LocationID string `json:"locationID"`

		// Created is the time of the player's creation.
		// in: body
		Created arcade.Timestamp `json:"created"`

		// Updated is the time the player was last updated.
		// in: body
		Updated arcade.Timestamp `json:"updated"`
	}
)

// Register sets up the http handler for this service with the given router.
func (s PlayersService) Register(router *mux.Router) {
	r := router.PathPrefix(V1PlayersRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{playerID}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{playerID}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{playerID}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (PlayersService) Name() string {
	return "players"
}

// Shutdown is a no-op since there no long running processes for this service.
func (PlayersService) Shutdown() {}

// List handles a request to retrieve multiple players.
func (s PlayersService) List(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/players List
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
	filter, err := NewPlayersFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of players.
	aPlayers, err := s.Manager.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from arcade players, to local players.
	var players []Player
	for _, aPlayer := range aPlayers {
		players = append(players, TranslatePlayer(aPlayer))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressPlayers{Players: players})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve a player.
func (s PlayersService) Get(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/players/{playerID} Get
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
	playerID := mux.Vars(r)["playerID"]
	aPlayerID, err := uuid.Parse(playerID)
	if err != nil {
		err := fmt.Errorf("%w: invalid playerID, not a well formed uuid: '%s'", errors.ErrBadRequest, playerID)
		server.Response(ctx, w, err)
		return
	}

	// Request the player from the player manager.
	aPlayer, err := s.Manager.Get(ctx, arcade.PlayerID(aPlayerID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the player to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressPlayer{Player: TranslatePlayer(aPlayer)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create a player.
func (s PlayersService) Create(w http.ResponseWriter, r *http.Request) {
	// swagger:route POST /v1/players
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

	var ingressPlayer IngressPlayer
	err = json.Unmarshal(body, &ingressPlayer)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the player request to the player manager.
	aIngressPlayer, err := TranslateIngressPlayer(ingressPlayer)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	aPlayer, err := s.Manager.Create(ctx, aIngressPlayer)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned player for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressPlayer{Player: TranslatePlayer(aPlayer)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update a player.
func (s PlayersService) Update(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /v1/players/{playerID}
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
	playerID := mux.Vars(r)["playerID"]
	u, err := uuid.Parse(playerID)
	if err != nil {
		err := fmt.Errorf("%w: invalid playerID, not a well formed uuid: '%s'", errors.ErrBadRequest, playerID)
		server.Response(ctx, w, err)
		return
	}
	aPlayerID := arcade.PlayerID(u)

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

	// Populate the ingress player from the body.
	var ingressPlayer IngressPlayer
	err = json.Unmarshal(body, &ingressPlayer)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the player request.
	aIngressPlayer, err := TranslateIngressPlayer(ingressPlayer)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the player to the player manager.
	aPlayer, err := s.Manager.Update(ctx, aPlayerID, aIngressPlayer)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressPlayer{Player: TranslatePlayer(aPlayer)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove a player.
func (s PlayersService) Remove(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /v1/players/{playerID} Get
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
	playerID := mux.Vars(r)["playerID"]
	aPlayerID, err := uuid.Parse(playerID)
	if err != nil {
		err := fmt.Errorf("%w: invalid playerID, not a well formed uuid: '%s'", errors.ErrBadRequest, playerID)
		server.Response(ctx, w, err)
		return
	}

	// Send the playerID to the player manager for removal.
	err = s.Manager.Remove(ctx, arcade.PlayerID(aPlayerID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewPlayersFilter creates a PlayersFilter from the the given request's URL
// query parameters.
func NewPlayersFilter(r *http.Request) (arcade.PlayersFilter, error) {
	q := r.URL.Query()
	filter := arcade.PlayersFilter{
		Limit: arcade.DefaultPlayersFilterLimit,
	}

	if values := q["locationID"]; len(values) > 0 {
		locationID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.PlayersFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.LocationID = arcade.RoomID(locationID)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > arcade.MaxPlayersFilterLimit {
			return arcade.PlayersFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return arcade.PlayersFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	return filter, nil
}

// TranslatesIngressPlayer translates the player request from the http request to an arcade.PlayerRequest.
func TranslateIngressPlayer(l IngressPlayer) (arcade.IngressPlayer, error) {
	empty := arcade.IngressPlayer{}

	if l.Name == "" {
		return empty, fmt.Errorf("%w: empty player name", errors.ErrBadRequest)
	}
	if len(l.Name) > arcade.MaxPlayerNameLen {
		return empty, fmt.Errorf("%w: player name exceeds maximum length", errors.ErrBadRequest)
	}
	if l.Description == "" {
		return empty, fmt.Errorf("%w: empty player description", errors.ErrBadRequest)
	}
	if len(l.Description) > arcade.MaxPlayerDescriptionLen {
		return empty, fmt.Errorf("%w: player description exceeds maximum length", errors.ErrBadRequest)
	}
	homeID, err := uuid.Parse(l.HomeID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid homeID: '%s'", errors.ErrBadRequest, l.HomeID)
	}
	locID, err := uuid.Parse(l.LocationID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid locationID: '%s', %s", errors.ErrBadRequest, l.LocationID, err)
	}

	return arcade.IngressPlayer{
		Name:        l.Name,
		Description: l.Description,
		HomeID:      arcade.RoomID(homeID),
		LocationID:  arcade.RoomID(locID),
	}, nil
}

// TranslatePlayer translates an arcade player to a local player.
func TranslatePlayer(l *arcade.Player) Player {
	return Player{
		ID:          l.ID.String(),
		Name:        l.Name,
		Description: l.Description,
		HomeID:      l.HomeID.String(),
		LocationID:  l.LocationID.String(),
		Created:     l.Created,
		Updated:     l.Updated,
	}
}
