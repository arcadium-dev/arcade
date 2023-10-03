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
	V1RoomsRoute string = "/v1/rooms"
)

type (
	// RoomService services room related network requests.
	RoomsService struct {
		Manager RoomManager
	}

	// RoomManager defines the expected behavior of the room manager in the domain layer.
	RoomManager interface {
		List(ctx context.Context, filter arcade.RoomsFilter) ([]*arcade.Room, error)
		Get(ctx context.Context, roomID arcade.RoomID) (*arcade.Room, error)
		Create(ctx context.Context, ingressRoom arcade.IngressRoom) (*arcade.Room, error)
		Update(ctx context.Context, roomID arcade.RoomID, ingressRoom arcade.IngressRoom) (*arcade.Room, error)
		Remove(ctx context.Context, roomID arcade.RoomID) error
	}

	// IngressRoom is used to request a room be created or updated.
	//
	// swagger:parameters RoomCreate RoomUpdate
	IngressRoom struct {
		// Name is the name of the room.
		// in: body
		// minimum length: 1
		// maximum length: 256
		Name string `json:"name"`

		// Description is the description of the room.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		Description string `json:"description"`

		// OwnerID is the ID of the owner of the room.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		OwnerID string `json:"ownerID"`

		// ParentID is the ID of the parent of the room.
		// in: body
		ParentID string `json:"parentID"`
	}

	// EgressRoom returns a room.
	EgressRoom struct {
		// Room returns the information about a room.
		// in: body
		Room Room `json:"room"`
	}

	// RoomsResponse returns multiple rooms.
	EgressRooms struct {
		// Rooms returns the information about multiple rooms.
		// in: body
		Rooms []Room `json:"rooms"`
	}

	// Room holds a room's information, and is sent in a response.
	//
	// swagger:parameter
	Room struct {
		// ID is the room identifier.
		// in: body
		ID string `json:"id"`

		// Name is the room name.
		// in: body
		Name string `json:"name"`

		// Description is the room description.
		// in: body
		Description string `json:"description"`

		// OwnerID is the PlayerID of the room owner.
		// in:body
		OwnerID string `json:"ownerID"`

		// ParentID is the RoomID of the room's parent.
		// in: body
		ParentID string `json:"parentID"`

		// Created is the time of the room's creation.
		// in: body
		Created arcade.Timestamp `json:"created"`

		// Updated is the time the room was last updated.
		// in: body
		Updated arcade.Timestamp `json:"updated"`
	}
)

// Register sets up the http handler for this service with the given router.
func (s RoomsService) Register(router *mux.Router) {
	r := router.PathPrefix(V1RoomsRoute).Subrouter()
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
	// swagger:route GET /v1/rooms List
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
	filter, err := NewRoomsFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of rooms.
	aRooms, err := s.Manager.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from arcade rooms, to local rooms.
	var rooms []Room
	for _, aRoom := range aRooms {
		rooms = append(rooms, TranslateRoom(aRoom))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressRooms{Rooms: rooms})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve a room.
func (s RoomsService) Get(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/rooms/{roomID} Get
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
	roomID := mux.Vars(r)["roomID"]
	aRoomID, err := uuid.Parse(roomID)
	if err != nil {
		err := fmt.Errorf("%w: invalid roomID, not a well formed uuid: '%s'", errors.ErrBadRequest, roomID)
		server.Response(ctx, w, err)
		return
	}

	// Request the room from the room manager.
	aRoom, err := s.Manager.Get(ctx, arcade.RoomID(aRoomID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the room to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressRoom{Room: TranslateRoom(aRoom)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create a room.
func (s RoomsService) Create(w http.ResponseWriter, r *http.Request) {
	// swagger:route POST /v1/rooms
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

	var ingressRoom IngressRoom
	err = json.Unmarshal(body, &ingressRoom)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the room request to the room manager.
	aIngressRoom, err := TranslateIngressRoom(ingressRoom)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	aRoom, err := s.Manager.Create(ctx, aIngressRoom)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned room for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressRoom{Room: TranslateRoom(aRoom)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update a room.
func (s RoomsService) Update(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /v1/rooms/{roomID}
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
	roomID := mux.Vars(r)["roomID"]
	u, err := uuid.Parse(roomID)
	if err != nil {
		err := fmt.Errorf("%w: invalid roomID, not a well formed uuid: '%s'", errors.ErrBadRequest, roomID)
		server.Response(ctx, w, err)
		return
	}
	aRoomID := arcade.RoomID(u)

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

	// Populate the ingress room from the body.
	var ingressRoom IngressRoom
	err = json.Unmarshal(body, &ingressRoom)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the room request.
	aIngressRoom, err := TranslateIngressRoom(ingressRoom)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the room to the room manager.
	aRoom, err := s.Manager.Update(ctx, aRoomID, aIngressRoom)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressRoom{Room: TranslateRoom(aRoom)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove a room.
func (s RoomsService) Remove(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /v1/rooms/{roomID} Get
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
	roomID := mux.Vars(r)["roomID"]
	aRoomID, err := uuid.Parse(roomID)
	if err != nil {
		err := fmt.Errorf("%w: invalid roomID, not a well formed uuid: '%s'", errors.ErrBadRequest, roomID)
		server.Response(ctx, w, err)
		return
	}

	// Send the roomID to the room manager for removal.
	err = s.Manager.Remove(ctx, arcade.RoomID(aRoomID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewRoomsFilter creates a RoomsFilter from the the given request's URL
// query parameters
func NewRoomsFilter(r *http.Request) (arcade.RoomsFilter, error) {
	q := r.URL.Query()
	filter := arcade.RoomsFilter{
		Limit: arcade.DefaultRoomsFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.RoomsFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = arcade.PlayerID(ownerID)
	}

	if values := q["parentID"]; len(values) > 0 {
		parentID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.RoomsFilter{}, fmt.Errorf("%w: invalid parentID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.ParentID = arcade.RoomID(parentID)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > arcade.MaxRoomsFilterLimit {
			return arcade.RoomsFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return arcade.RoomsFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	return filter, nil
}

// IngressRoomtranslates the room request from the http request to an arcade.RoomRequest.
func TranslateIngressRoom(l IngressRoom) (arcade.IngressRoom, error) {
	empty := arcade.IngressRoom{}

	if l.Name == "" {
		return empty, fmt.Errorf("%w: empty room name", errors.ErrBadRequest)
	}
	if len(l.Name) > arcade.MaxRoomNameLen {
		return empty, fmt.Errorf("%w: room name exceeds maximum length", errors.ErrBadRequest)
	}
	if l.Description == "" {
		return empty, fmt.Errorf("%w: empty room description", errors.ErrBadRequest)
	}
	if len(l.Description) > arcade.MaxRoomDescriptionLen {
		return empty, fmt.Errorf("%w: room description exceeds maximum length", errors.ErrBadRequest)
	}
	ownerID, err := uuid.Parse(l.OwnerID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid ownerID: '%s'", errors.ErrBadRequest, l.OwnerID)
	}
	parentID, err := uuid.Parse(l.ParentID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid parentID: '%s', %s", errors.ErrBadRequest, l.ParentID, err)
	}

	return arcade.IngressRoom{
		Name:        l.Name,
		Description: l.Description,
		OwnerID:     arcade.PlayerID(ownerID),
		ParentID:    arcade.RoomID(parentID),
	}, nil
}

// TranslateRoom translates an arcade room to a local room.
func TranslateRoom(l *arcade.Room) Room {
	return Room{
		ID:          l.ID.String(),
		Name:        l.Name,
		Description: l.Description,
		OwnerID:     l.OwnerID.String(),
		ParentID:    l.ParentID.String(),
		Created:     l.Created,
		Updated:     l.Updated,
	}
}
