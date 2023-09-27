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
	RoomsRoute string = "/rooms"
)

type (
	// Rooms is used to manage the room assets.
	RoomsService struct {
		Storage arcade.RoomsStorage
	}
)

// Register sets up the http handler for this service with the given router.
func (s RoomsService) Register(router *mux.Router) {
	r := router.PathPrefix(RoomsRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{roomID}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{roomID}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{roomID}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (RoomsService) Name() string {
	return "rooms"
}

// Shutdown is a no-op since there no long running processes for this service.
func (RoomsService) Shutdown() {}

// List handles a request to retrieve multiple rooms.
func (s RoomsService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create the filter.
	filter, err := arcade.NewRoomsFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of rooms.
	rooms, err := s.Storage.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.NewRoomsResponse(rooms))
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve a room.
func (s RoomsService) Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomID := params["roomID"]

	ctx := r.Context()

	room, err := s.Storage.Get(ctx, roomID)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.RoomResponse{Data: room})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to retrieve a room.
func (s RoomsService) Create(w http.ResponseWriter, r *http.Request) {
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

	var req arcade.RoomRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	room, err := s.Storage.Create(ctx, req)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.RoomResponse{Data: room})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update a room.
func (s RoomsService) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	roomID := params["roomID"]

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

	var req arcade.RoomRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	room, err := s.Storage.Update(ctx, roomID, req)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.RoomResponse{Data: room})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove a room.
func (s RoomsService) Remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	roomID := params["roomID"]

	err := s.Storage.Remove(ctx, roomID)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
