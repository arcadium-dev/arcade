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

package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	cerrors "arcadium.dev/core/errors"
	chttp "arcadium.dev/core/http"

	"arcadium.dev/arcade"
)

const (
	roomsRoute string = "/rooms"
)

type (
	// Rooms is used to manage the room assets.
	RoomsService struct {
		Storage arcade.RoomStorage
	}
)

// Register sets up the http handler for this service with the given router.
func (s RoomsService) Register(router *mux.Router) {
	r := router.PathPrefix(roomsRoute).Subrouter()
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

// Shutdown is a no-op since there no long running processes for this service... yet.
func (RoomsService) Shutdown() {}

func (s RoomsService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: parse query params

	// Read list of rooms.
	rooms, err := s.Storage.List(ctx, arcade.RoomsFilter{})
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.NewRoomsResponse(rooms))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (s RoomsService) Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomID := params["roomID"]

	ctx := r.Context()

	room, err := s.Storage.Get(ctx, roomID)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.RoomResponse{Data: room})
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (s RoomsService) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid json: a json encoded body is required", cerrors.ErrInvalidArgument,
		))
		return
	}

	var req arcade.RoomRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}

	room, err := s.Storage.Create(ctx, req)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.RoomResponse{Data: room})
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (s RoomsService) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	roomID := params["roomID"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid json: a json encoded body is required", cerrors.ErrInvalidArgument,
		))
		return
	}

	var req arcade.RoomRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}

	room, err := s.Storage.Update(ctx, roomID, req)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.RoomResponse{Data: room})
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (s RoomsService) Remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	roomID := params["roomID"]

	err := s.Storage.Remove(ctx, roomID)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}