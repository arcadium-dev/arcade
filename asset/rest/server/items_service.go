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

package server // import "arcadium.dev/arcade/asset/rest/server"

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
	"github.com/rs/zerolog"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
)

const (
	V1ItemRoute string = "/v1/item"
)

type (
	// ItemService services item related network requests.
	ItemsService struct {
		Storage ItemStorage
	}

	// ItemStorage defines the expected behavior of the item manager in the domain layer.
	ItemStorage interface {
		List(context.Context, asset.ItemFilter) ([]*asset.Item, error)
		Get(context.Context, asset.ItemID) (*asset.Item, error)
		Create(context.Context, asset.ItemCreate) (*asset.Item, error)
		Update(context.Context, asset.ItemID, asset.ItemUpdate) (*asset.Item, error)
		Remove(context.Context, asset.ItemID) error
	}
)

// Register sets up the http handler for this service with the given router.
func (s ItemsService) Register(router *mux.Router) {
	r := router.PathPrefix(V1ItemRoute).Subrouter()
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
	filter, err := NewItemFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of items.
	aItems, err := s.Storage.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from asset items, to network items.
	items := make([]rest.Item, 0)
	for _, aItem := range aItems {
		items = append(items, TranslateItem(aItem))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.ItemsResponse{Items: items})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode item list response, error %s", err)
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
	item, err := s.Storage.Get(ctx, asset.ItemID(itemID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the item to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.ItemResponse{Item: TranslateItem(item)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode item get response, error %s", err)
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

	var createReq rest.ItemCreateRequest
	err = json.Unmarshal(body, &createReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the item request to the item manager.
	change, err := TranslateItemRequest(createReq.ItemRequest)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	item, err := s.Storage.Create(ctx, asset.ItemCreate{ItemChange: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned item for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(rest.ItemResponse{Item: TranslateItem(item)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode item create response, error %s", err)
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
	var updateReq rest.ItemUpdateRequest
	err = json.Unmarshal(body, &updateReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the item request.
	change, err := TranslateItemRequest(updateReq.ItemRequest)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the item to the item manager.
	item, err := s.Storage.Update(ctx, asset.ItemID(itemID), asset.ItemUpdate{ItemChange: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.ItemResponse{Item: TranslateItem(item)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode item update response, error %s", err)
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
	err = s.Storage.Remove(ctx, asset.ItemID(itemID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewItemFilter creates an asset items filter from the the given request's query parameters.
func NewItemFilter(r *http.Request) (asset.ItemFilter, error) {
	q := r.URL.Query()
	filter := asset.ItemFilter{
		Limit: asset.DefaultItemFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return asset.ItemFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = asset.PlayerID(ownerID)
	}

	var (
		locationUUID *uuid.UUID
	)
	if values := q["locationID"]; len(values) > 0 {
		u, err := uuid.Parse(values[0])
		locationUUID = &u
		if err != nil {
			return asset.ItemFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
	}

	if locationUUID != nil {
		var location asset.ItemLocationID
		if values := q["locationType"]; len(values) > 0 {
			switch strings.ToLower(values[0]) {
			case "room":
				location = asset.RoomID(*locationUUID)
			case "player":
				location = asset.PlayerID(*locationUUID)
			case "item":
				location = asset.ItemID(*locationUUID)
			default:
				return asset.ItemFilter{}, fmt.Errorf("%w: invalid locationType query parameter: '%s'", errors.ErrBadRequest, values[0])
			}
		} else {
			return asset.ItemFilter{}, fmt.Errorf("%w: locationType required when locationID is set", errors.ErrBadRequest)
		}
		filter.LocationID = location
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return asset.ItemFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > asset.MaxItemFilterLimit {
			return asset.ItemFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	return filter, nil
}

// TranslateItemRequest translates a network asset item request to an asset item request.
func TranslateItemRequest(i rest.ItemRequest) (asset.ItemChange, error) {
	empty := asset.ItemChange{}

	if i.Name == "" {
		return empty, fmt.Errorf("%w: empty item name", errors.ErrBadRequest)
	}
	if len(i.Name) > asset.MaxItemNameLen {
		return empty, fmt.Errorf("%w: item name exceeds maximum length", errors.ErrBadRequest)
	}
	if i.Description == "" {
		return empty, fmt.Errorf("%w: empty item description", errors.ErrBadRequest)
	}
	if len(i.Description) > asset.MaxItemDescriptionLen {
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

	itemReq := asset.ItemChange{
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     asset.PlayerID(ownerID),
	}

	switch t {
	case "room":
		itemReq.LocationID = asset.RoomID(locID)
	case "player":
		itemReq.LocationID = asset.PlayerID(locID)
	case "item":
		itemReq.LocationID = asset.ItemID(locID)
	}

	return itemReq, nil
}

// TranslateItem translates an asset item to a network item.
func TranslateItem(i *asset.Item) rest.Item {
	return rest.Item{
		ID:          i.ID.String(),
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     i.OwnerID.String(),
		LocationID: rest.ItemLocationID{
			ID:   i.LocationID.ID().String(),
			Type: i.LocationID.Type().String(),
		},
		Created: i.Created,
		Updated: i.Updated,
	}
}
