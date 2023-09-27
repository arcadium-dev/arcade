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

package service // import "arcadium.dev/arcade/service"

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade"
)

const (
	PlayersRoute string = "/players"
)

type (
	// Players is used to manage the player assets.
	PlayersService struct {
		Storage arcade.PlayersStorage
	}
)

// Register sets up the http handler for this service with the given router.
func (s PlayersService) Register(router *mux.Router) {
	r := router.PathPrefix(PlayersRoute).Subrouter()
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
	ctx := r.Context()

	// Create the filter.
	filter, err := arcade.NewPlayersFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of players.
	players, err := s.Storage.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.NewPlayersResponse(players))
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve a player.
func (s PlayersService) Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	playerID := params["playerID"]

	ctx := r.Context()

	player, err := s.Storage.Get(ctx, playerID)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.PlayerResponse{Data: player})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create a player.
func (s PlayersService) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request: %s", errors.ErrBadRequest, err,
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

	var req arcade.PlayerRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	player, err := s.Storage.Create(ctx, req)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.PlayerResponse{Data: player})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update a player.
func (s PlayersService) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	playerID := params["playerID"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request: %s", errors.ErrBadRequest, err,
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

	var req arcade.PlayerRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	player, err := s.Storage.Update(ctx, playerID, req)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.PlayerResponse{Data: player})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove a player.
func (s PlayersService) Remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	playerID := params["playerID"]

	err := s.Storage.Remove(ctx, playerID)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
