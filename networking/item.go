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
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade"
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
		List(ctx context.Context, filter arcade.ItemsFilter) ([]*arcade.Item, error)
		Get(ctx context.Context, itemID arcade.ItemID) (*arcade.Item, error)
		Create(ctx context.Context, ingressItem arcade.IngressItem) (*arcade.Item, error)
		Update(ctx context.Context, itemID arcade.ItemID, ingressItem arcade.IngressItem) (*arcade.Item, error)
		Remove(ctx context.Context, itemID arcade.ItemID) error
	}

	// IngressItem is used to request an item be created or updated.
	//
	// swagger:parameters ItemCreate ItemUpdate
	IngressItem struct {
		// Name is the name of the item.
		// in: body
		// minimum length: 1
		// maximum length: 256
		Name string `json:"name"`

		// Description is the description of the item.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		Description string `json:"description"`

		// OwnerID is the ID of the owner of the item.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		OwnerID string `json:"ownerID"`

		// LocationID is the ID of the location of the item.
		// in: body
		LocationID ItemLocationID `json:"locationID"`
	}

	// EgressItem returns an item.
	EgressItem struct {
		// Item returns the information about an item.
		// in: body
		Item Item `json:"item"`
	}

	// ItemsResponse returns multiple items.
	EgressItems struct {
		// Items returns the information about multiple items.
		// in: body
		Items []Item `json:"items"`
	}

	// Item holds an item's information, and is sent in a response.
	//
	// swagger:parameter
	Item struct {
		// ID is the item identifier.
		// in: body
		ID string `json:"id"`

		// Name is the item name.
		// in: body
		Name string `json:"name"`

		// Description is the item description.
		// in: body
		Description string `json:"description"`

		// OwnerID is the PlayerID of the item owner.
		// in:body
		OwnerID string `json:"ownerID"`

		// LocationID is the LocationID of the item's location.
		// in: body
		LocationID ItemLocationID `json:"locationID"`

		// Created is the time of the item's creation.
		// in: body
		Created arcade.Timestamp `json:"created"`

		// Updated is the time the item was last updated.
		// in: body
		Updated arcade.Timestamp `json:"updated"`
	}

	// ItemLocationID holds the locationID of the item, and the type of location.
	ItemLocationID struct {
		// ID is the location identifier. This can correspond the the ID of a room, player or item.
		// in: body
		ID string `json:"id"`

		// Type is the type of location. This can be "room", "player" or "item".
		Type string `json:"type"`
	}
)

// Register sets up the http handler for this service with the given router.
func (s ItemsService) Register(router *mux.Router) {
	r := router.PathPrefix(V1ItemsRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{itemID}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{itemID}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{itemID}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (ItemsService) Name() string {
	return "items"
}

// Shutdown is a no-op since there no long running processes for this service.
func (ItemsService) Shutdown() {}

// List handles a request to retrieve multiple items.
func (s ItemsService) List(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/items List
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

	// Translate from arcade items, to local items.
	var items []Item
	for _, aItem := range aItems {
		items = append(items, TranslateItem(aItem))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressItems{Items: items})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve an item.
func (s ItemsService) Get(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/items/{itemID} Get
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
	itemID := mux.Vars(r)["itemID"]
	aItemID, err := uuid.Parse(itemID)
	if err != nil {
		err := fmt.Errorf("%w: invalid itemID, not a well formed uuid: '%s'", errors.ErrBadRequest, itemID)
		server.Response(ctx, w, err)
		return
	}

	// Request the item from the item manager.
	aItem, err := s.Manager.Get(ctx, arcade.ItemID(aItemID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the item to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressItem{Item: TranslateItem(aItem)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create an item.
func (s ItemsService) Create(w http.ResponseWriter, r *http.Request) {
	// swagger:route POST /v1/items
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

	var ingressItem IngressItem
	err = json.Unmarshal(body, &ingressItem)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the item request to the item manager.
	aIngressItem, err := TranslateIngressItem(ingressItem)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	aItem, err := s.Manager.Create(ctx, aIngressItem)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned item for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressItem{Item: TranslateItem(aItem)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update an item.
func (s ItemsService) Update(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /v1/items/{itemID}
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
	itemID := mux.Vars(r)["itemID"]
	u, err := uuid.Parse(itemID)
	if err != nil {
		err := fmt.Errorf("%w: invalid itemID, not a well formed uuid: '%s'", errors.ErrBadRequest, itemID)
		server.Response(ctx, w, err)
		return
	}
	aItemID := arcade.ItemID(u)

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

	// Populate the ingress item from the body.
	var ingressItem IngressItem
	err = json.Unmarshal(body, &ingressItem)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the item request.
	aIngressItem, err := TranslateIngressItem(ingressItem)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the item to the item manager.
	aItem, err := s.Manager.Update(ctx, aItemID, aIngressItem)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressItem{Item: TranslateItem(aItem)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove an item.
func (s ItemsService) Remove(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /v1/items/{itemID} Get
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
	itemID := mux.Vars(r)["itemID"]
	aItemID, err := uuid.Parse(itemID)
	if err != nil {
		err := fmt.Errorf("%w: invalid itemID, not a well formed uuid: '%s'", errors.ErrBadRequest, itemID)
		server.Response(ctx, w, err)
		return
	}

	// Send the itemID to the item manager for removal.
	err = s.Manager.Remove(ctx, arcade.ItemID(aItemID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewItemsFilter creates an ItemFilter from the the given request's URL
// query parameters.
func NewItemsFilter(r *http.Request) (arcade.ItemsFilter, error) {
	q := r.URL.Query()
	filter := arcade.ItemsFilter{
		Limit: arcade.DefaultItemsFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = arcade.PlayerID(ownerID)
	}

	var (
		locationUUID *uuid.UUID
	)
	if values := q["locationID"]; len(values) > 0 {
		u, err := uuid.Parse(values[0])
		locationUUID = &u
		if err != nil {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
	}

	if locationUUID != nil {
		var location arcade.ItemLocationID
		if values := q["locationType"]; len(values) > 0 {
			switch strings.ToLower(values[0]) {
			case "room":
				location = arcade.RoomID(*locationUUID)
			case "player":
				location = arcade.PlayerID(*locationUUID)
			case "item":
				location = arcade.ItemID(*locationUUID)
			default:
				return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid locationType query parameter: '%s'", errors.ErrBadRequest, values[0])
			}
		} else {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: locationType required when locationID is set", errors.ErrBadRequest)
		}
		filter.LocationID = location
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > arcade.MaxItemsFilterLimit {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	return filter, nil
}

// IngressItemtranslates the item request from the http request to an arcade.ItemRequest.
func TranslateIngressItem(i IngressItem) (arcade.IngressItem, error) {
	empty := arcade.IngressItem{}

	if i.Name == "" {
		return empty, fmt.Errorf("%w: empty item name", errors.ErrBadRequest)
	}
	if len(i.Name) > arcade.MaxItemNameLen {
		return empty, fmt.Errorf("%w: item name exceeds maximum length", errors.ErrBadRequest)
	}
	if i.Description == "" {
		return empty, fmt.Errorf("%w: empty item description", errors.ErrBadRequest)
	}
	if len(i.Description) > arcade.MaxItemDescriptionLen {
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

	itemReq := arcade.IngressItem{
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     arcade.PlayerID(ownerID),
	}

	switch t {
	case "room":
		itemReq.LocationID = arcade.RoomID(locID)
	case "player":
		itemReq.LocationID = arcade.PlayerID(locID)
	case "item":
		itemReq.LocationID = arcade.ItemID(locID)
	}

	return itemReq, nil
}

// TranslateItem translates an arcade item to a local item.
func TranslateItem(i *arcade.Item) Item {
	return Item{
		ID:          i.ID.String(),
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     i.OwnerID.String(),
		LocationID: ItemLocationID{
			ID:   i.LocationID.ID().String(),
			Type: i.LocationID.Type().String(),
		},
		Created: i.Created,
		Updated: i.Updated,
	}
}
