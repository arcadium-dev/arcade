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
	"github.com/rs/zerolog"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
)

const (
	V1RoomRoute string = "/v1/room"
)

type (
	// RoomService services room related network requests.
	RoomsService struct {
		Storage RoomStorage
	}

	// RoomStorage defines the expected behavior of the room manager in the domain layer.
	RoomStorage interface {
		List(context.Context, asset.RoomFilter) ([]*asset.Room, error)
		Get(context.Context, asset.RoomID) (*asset.Room, error)
		Create(context.Context, asset.RoomCreate) (*asset.Room, error)
		Update(context.Context, asset.RoomID, asset.RoomUpdate) (*asset.Room, error)
		Remove(context.Context, asset.RoomID) error
	}
)

// Register sets up the http handler for this service with the given router.
func (s RoomsService) Register(router *mux.Router) {
	r := router.PathPrefix(V1RoomRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{id}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{id}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{id}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (RoomsService) Name() string {
	return "rooms"
}

// Shutdown is a no-op since there no long running processes for this service.
func (RoomsService) Shutdown(context.Context) {}

// List handles a request to retrieve multiple rooms.
func (s RoomsService) List(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/rooms RoomList
	//
	// List returns a list of rooms.
	//
	// Produces: application/json
	//
	// Parameters:
	//   + name ownerID
	//     in: query
	//   + name: parentID
	//     in: query
	//   + name: offset
	//     in: query
	//   + name: limit
	//     in: query
	//
	// Responses:
	//  200: RoomResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Create a filter from the quesry parameters.
	filter, err := NewRoomFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of rooms.
	aRooms, err := s.Storage.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from asset rooms, to network rooms.
	rooms := make([]rest.Room, 0)
	for _, aRoom := range aRooms {
		rooms = append(rooms, TranslateRoom(aRoom))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.RoomsResponse{Rooms: rooms})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode room list response, error %s", err)
		return
	}
}

// Get handles a request to retrieve a room.
func (s RoomsService) Get(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/rooms/{roomID} RoomGet
	//
	// Get returns a room.
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: room ID
	//     required: true
	//
	// Responses:
	//  200: RoomResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the roomID from the uri.
	id := mux.Vars(r)["id"]
	roomID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid room id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Request the room from the room manager.
	room, err := s.Storage.Get(ctx, asset.RoomID(roomID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the room to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.RoomResponse{Room: TranslateRoom(room)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode room get response, error %s", err)
		return
	}
}

// Create handles a request to create a room.
func (s RoomsService) Create(w http.ResponseWriter, r *http.Request) {
	// swagger:route POST /v1/rooms RoomCreate
	//
	// Create will create a new room based on the room request in the body of the
	// request.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Responses:
	//  200: RoomResponse
	//  400: ResponseError
	//  409: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the room request from the body of the request.
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

	var createReq rest.RoomCreateRequest
	err = json.Unmarshal(body, &createReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the room request to the room manager.
	change, err := TranslateRoomRequest(createReq.RoomRequest)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	room, err := s.Storage.Create(ctx, asset.RoomCreate{RoomChange: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned room for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(rest.RoomResponse{Room: TranslateRoom(room)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode room create response, error %s", err)
		return
	}
}

// Update handles a request to update a room.
func (s RoomsService) Update(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /v1/rooms/{id} RoomUpdate
	//
	// Update will update room based on the roomID and the room\ request in the
	// body of the request.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: room ID
	//     required: true
	//
	// Responses:
	//  200: RoomResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Grab the roomID from the uri.
	id := mux.Vars(r)["id"]
	roomID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid room id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
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

	// Populate the network room from the body.
	var updateReq rest.RoomUpdateRequest
	err = json.Unmarshal(body, &updateReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the room request.
	change, err := TranslateRoomRequest(updateReq.RoomRequest)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the room to the room manager.
	room, err := s.Storage.Update(ctx, asset.RoomID(roomID), asset.RoomUpdate{RoomChange: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.RoomResponse{Room: TranslateRoom(room)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode room update response, error %s", err)
		return
	}
}

// Remove handles a request to remove a room.
func (s RoomsService) Remove(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /v1/rooms/{id} RoomRemove
	//
	// Remove deletes the room.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: room ID
	//     required: true
	//
	// Responses:
	//  200: RoomResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the roomID from the uri.
	id := mux.Vars(r)["id"]
	roomID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid room id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Send the roomID to the room manager for removal.
	err = s.Storage.Remove(ctx, asset.RoomID(roomID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewRoomFilter creates an asset rooms filter from the given request's URL query parameters.
func NewRoomFilter(r *http.Request) (asset.RoomFilter, error) {
	q := r.URL.Query()
	filter := asset.RoomFilter{
		Limit: asset.DefaultRoomFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return asset.RoomFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = asset.PlayerID(ownerID)
	}

	if values := q["parentID"]; len(values) > 0 {
		parentID, err := uuid.Parse(values[0])
		if err != nil {
			return asset.RoomFilter{}, fmt.Errorf("%w: invalid parentID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.ParentID = asset.RoomID(parentID)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > asset.MaxRoomFilterLimit {
			return asset.RoomFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return asset.RoomFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	return filter, nil
}

// TranslateRoomRequest translates a network room request to an asset room request.
func TranslateRoomRequest(r rest.RoomRequest) (asset.RoomChange, error) {
	empty := asset.RoomChange{}

	if r.Name == "" {
		return empty, fmt.Errorf("%w: empty room name", errors.ErrBadRequest)
	}
	if len(r.Name) > asset.MaxRoomNameLen {
		return empty, fmt.Errorf("%w: room name exceeds maximum length", errors.ErrBadRequest)
	}
	if r.Description == "" {
		return empty, fmt.Errorf("%w: empty room description", errors.ErrBadRequest)
	}
	if len(r.Description) > asset.MaxRoomDescriptionLen {
		return empty, fmt.Errorf("%w: room description exceeds maximum length", errors.ErrBadRequest)
	}
	ownerID, err := uuid.Parse(r.OwnerID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid ownerID: '%s'", errors.ErrBadRequest, r.OwnerID)
	}
	parentID, err := uuid.Parse(r.ParentID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid parentID: '%s', %s", errors.ErrBadRequest, r.ParentID, err)
	}

	return asset.RoomChange{
		Name:        r.Name,
		Description: r.Description,
		OwnerID:     asset.PlayerID(ownerID),
		ParentID:    asset.RoomID(parentID),
	}, nil
}

// TranslateRoom translates an asset room to a network room.
func TranslateRoom(r *asset.Room) rest.Room {
	return rest.Room{
		ID:          r.ID.String(),
		Name:        r.Name,
		Description: r.Description,
		OwnerID:     r.OwnerID.String(),
		ParentID:    r.ParentID.String(),
		Created:     r.Created,
		Updated:     r.Updated,
	}
}
