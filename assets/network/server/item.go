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

package server // import "arcadium.dev/arcade/assets/network/server"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/arcade/assets/network"
)

const (
	V1ItemsRoute string = "/v1/items"
)

type (
	// ItemService services item related network requests.
	ItemsService struct {
		Manager ItemManager
	}

	// ItemManager defines the expected behavior of the item manager in the domain layer.
	ItemManager interface {
		List(context.Context, assets.ItemsFilter) ([]*assets.Item, error)
		Get(context.Context, assets.ItemID) (*assets.Item, error)
		Create(context.Context, assets.ItemCreateRequest) (*assets.Item, error)
		Update(context.Context, assets.ItemID, assets.ItemUpdateRequest) (*assets.Item, error)
		Remove(context.Context, assets.ItemID) error
	}
)

// Register sets up the http handler for this service with the given router.
func (s ItemsService) Register(router *mux.Router) {
	r := router.PathPrefix(V1ItemsRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{id}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{id}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{id}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (ItemsService) Name() string {
	return "items"
}

// Shutdown is a no-op since there no long running processes for this service.
func (ItemsService) Shutdown() {}

// List handles a request to retrieve multiple items.
func (s ItemsService) List(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/items ItemList
	//
	// List returns a list of items.
	//
	// Produces: application/json
	//
	// Parameters:
	//   + name ownerID
	//     in: query
	//   + name: locationID
	//     in: query
	//   + name: locationType
	//     in: query
	//   + name: offset
	//     in: query
	//   + name: limit
	//     in: query
	//
	// Responses:
	//  200: ItemResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Create a filter from the quesry parameters.
	filter, err := NewItemsFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of items.
	aItems, err := s.Manager.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from assets items, to network items.
	var items []network.Item
	for _, aItem := range aItems {
		items = append(items, TranslateItem(aItem))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(network.ItemsResponse{Items: items})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve an item.
func (s ItemsService) Get(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/items/{itemID} ItemGet
	//
	// Get returns an item.
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: item ID
	//     required: true
	//
	// Responses:
	//  200: ItemResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the itemID from the uri.
	id := mux.Vars(r)["id"]
	itemID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid item id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Request the item from the item manager.
	item, err := s.Manager.Get(ctx, assets.ItemID(itemID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the item to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(network.ItemResponse{Item: TranslateItem(item)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create an item.
func (s ItemsService) Create(w http.ResponseWriter, r *http.Request) {
	// swagger:route POST /v1/items ItemCreate
	//
	// Create will create a new item based on the item request in the body of the
	// request.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Responses:
	//  200: ItemResponse
	//  400: ResponseError
	//  409: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the item request from the body of the request.
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

	var createReq network.ItemCreateRequest
	err = json.Unmarshal(body, &createReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the item request to the item manager.
	req, err := TranslateItemRequest(createReq)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	item, err := s.Manager.Create(ctx, req)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned item for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(network.ItemResponse{Item: TranslateItem(item)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update an item.
func (s ItemsService) Update(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /v1/items/{id} ItemUpdate
	//
	// Update will update item based on the itemID and the item\ request in the
	// body of the request.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: item ID
	//     required: true
	//
	// Responses:
	//  200: ItemResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Grab the itemID from the uri.
	id := mux.Vars(r)["id"]
	itemID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid item id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
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

	// Populate the network item from the body.
	var updateReq network.ItemUpdateRequest
	err = json.Unmarshal(body, &updateReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the item request.
	req, err := TranslateItemRequest(updateReq)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the item to the item manager.
	item, err := s.Manager.Update(ctx, assets.ItemID(itemID), req)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(network.ItemResponse{Item: TranslateItem(item)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove an item.
func (s ItemsService) Remove(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /v1/items/{id} ItemRemove
	//
	// Remove deletes the item.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: item ID
	//     required: true
	//
	// Responses:
	//  200: ItemResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the itemID from the uri.
	id := mux.Vars(r)["id"]
	itemID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid item id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Send the itemID to the item manager for removal.
	err = s.Manager.Remove(ctx, assets.ItemID(itemID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewItemsFilter creates an assets items filter from the the given request's query parameters.
func NewItemsFilter(r *http.Request) (assets.ItemsFilter, error) {
	q := r.URL.Query()
	filter := assets.ItemsFilter{
		Limit: assets.DefaultItemsFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return assets.ItemsFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = assets.PlayerID(ownerID)
	}

	var (
		locationUUID *uuid.UUID
	)
	if values := q["locationID"]; len(values) > 0 {
		u, err := uuid.Parse(values[0])
		locationUUID = &u
		if err != nil {
			return assets.ItemsFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
	}

	if locationUUID != nil {
		var location assets.ItemLocationID
		if values := q["locationType"]; len(values) > 0 {
			switch strings.ToLower(values[0]) {
			case "room":
				location = assets.RoomID(*locationUUID)
			case "player":
				location = assets.PlayerID(*locationUUID)
			case "item":
				location = assets.ItemID(*locationUUID)
			default:
				return assets.ItemsFilter{}, fmt.Errorf("%w: invalid locationType query parameter: '%s'", errors.ErrBadRequest, values[0])
			}
		} else {
			return assets.ItemsFilter{}, fmt.Errorf("%w: locationType required when locationID is set", errors.ErrBadRequest)
		}
		filter.LocationID = location
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return assets.ItemsFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > assets.MaxItemsFilterLimit {
			return assets.ItemsFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	return filter, nil
}

// TranslateItemRequest translates a network asset item request to an assets item request.
func TranslateItemRequest(i network.ItemRequest) (assets.ItemRequest, error) {
	empty := assets.ItemRequest{}

	if i.Name == "" {
		return empty, fmt.Errorf("%w: empty item name", errors.ErrBadRequest)
	}
	if len(i.Name) > assets.MaxItemNameLen {
		return empty, fmt.Errorf("%w: item name exceeds maximum length", errors.ErrBadRequest)
	}
	if i.Description == "" {
		return empty, fmt.Errorf("%w: empty item description", errors.ErrBadRequest)
	}
	if len(i.Description) > assets.MaxItemDescriptionLen {
		return empty, fmt.Errorf("%w: item description exceeds maximum length", errors.ErrBadRequest)
	}
	ownerID, err := uuid.Parse(i.OwnerID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid ownerID: '%s'", errors.ErrBadRequest, i.OwnerID)
	}
	locID, err := uuid.Parse(i.LocationID.ID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid locationID.ID: '%s', %s", errors.ErrBadRequest, i.LocationID.ID, err)
	}
	t := strings.ToLower(i.LocationID.Type)
	if t != "room" && t != "player" && t != "item" {
		return empty, fmt.Errorf("%w: invalid locationID.Type: '%s'", errors.ErrBadRequest, i.LocationID.Type)
	}

	itemReq := assets.ItemRequest{
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     assets.PlayerID(ownerID),
	}

	switch t {
	case "room":
		itemReq.LocationID = assets.RoomID(locID)
	case "player":
		itemReq.LocationID = assets.PlayerID(locID)
	case "item":
		itemReq.LocationID = assets.ItemID(locID)
	}

	return itemReq, nil
}

// TranslateItem translates an assets item to a network item.
func TranslateItem(i *assets.Item) network.Item {
	return network.Item{
		ID:          i.ID.String(),
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     i.OwnerID.String(),
		LocationID: network.ItemLocationID{
			ID:   i.LocationID.ID().String(),
			Type: i.LocationID.Type().String(),
		},
		Created: i.Created,
		Updated: i.Updated,
	}
}
