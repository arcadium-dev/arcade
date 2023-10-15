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

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/arcade/assets/network/rest"
)

const (
	V1LinksRoute string = "/v1/links"
)

type (
	// LinkService services link related network requests.
	LinksService struct {
		Storage LinkStorage
	}

	// LinkStorage defines the expected behavior of the link manager in the domain layer.
	LinkStorage interface {
		List(context.Context, assets.LinkFilter) ([]*assets.Link, error)
		Get(context.Context, assets.LinkID) (*assets.Link, error)
		Create(context.Context, assets.LinkCreate) (*assets.Link, error)
		Update(context.Context, assets.LinkID, assets.LinkUpdate) (*assets.Link, error)
		Remove(context.Context, assets.LinkID) error
	}
)

// Register sets up the http handler for this service with the given router.
func (s LinksService) Register(router *mux.Router) {
	r := router.PathPrefix(V1LinksRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{id}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{id}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{id}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (LinksService) Name() string {
	return "links"
}

// Shutdown is a no-op since there no long running processes for this service.
func (LinksService) Shutdown() {}

// List handles a request to retrieve multiple links.
func (s LinksService) List(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/links LinkList
	//
	// List returns a list of links.
	//
	// Produces: application/json
	//
	// Parameters:
	//   + name ownerID
	//     in: query
	//   + name: locationID
	//     in: query
	//   + name: destinationID
	//     in: query
	//   + name: offset
	//     in: query
	//   + name: limit
	//     in: query
	//
	// Responses:
	//  200: LinkResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Create a filter from the quesry parameters.
	filter, err := NewLinkFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of links.
	aLinks, err := s.Storage.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from assets links, to network links.
	links := make([]rest.Link, 0)
	for _, aLink := range aLinks {
		links = append(links, TranslateLink(aLink))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.LinksResponse{Links: links})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve a link.
func (s LinksService) Get(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/links/{id} LinkGet
	//
	// Get returns a link.
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: link ID
	//     required: true
	//
	// Responses:
	//  200: LinkResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the linkID from the uri.
	id := mux.Vars(r)["id"]
	linkID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid link id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Request the link from the link manager.
	link, err := s.Storage.Get(ctx, assets.LinkID(linkID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the link to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.LinkResponse{Link: TranslateLink(link)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create a link.
func (s LinksService) Create(w http.ResponseWriter, r *http.Request) {
	// swagger:route POST /v1/links LinkCreate
	//
	// Create will create a new link based on the link request in the body of the
	// request.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Responses:
	//  200: LinkResponse
	//  400: ResponseError
	//  409: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the link request from the body of the request.
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

	var createReq rest.LinkCreateRequest
	err = json.Unmarshal(body, &createReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the link request to the link manager.
	change, err := TranslateLinkRequest(createReq.LinkRequest)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	link, err := s.Storage.Create(ctx, assets.LinkCreate{LinkChange: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned link for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.LinkResponse{Link: TranslateLink(link)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update a link.
func (s LinksService) Update(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /v1/links/{id} LinkUpdate
	//
	// Update will update link based on the linkID and the link request in the
	// body of the request.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: link ID
	//     required: true
	//
	// Responses:
	//  200: LinkResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Grab the linkID from the uri.
	id := mux.Vars(r)["id"]
	linkID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid link id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
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

	var updateReq rest.LinkUpdateRequest
	err = json.Unmarshal(body, &updateReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the link request.
	change, err := TranslateLinkRequest(updateReq.LinkRequest)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the link to the link manager.
	link, err := s.Storage.Update(ctx, assets.LinkID(linkID), assets.LinkUpdate{LinkChange: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(rest.LinkResponse{Link: TranslateLink(link)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove a link.
func (s LinksService) Remove(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /v1/links/{id} LinkRemove
	//
	// Remove deletes the link.
	//
	// Consumes: application/json
	//
	// Produces: application/json
	//
	// Parameters:
	// 	 + name: id
	//     in: path
	//     description: link ID
	//     required: true
	//
	// Responses:
	//  200: LinkResponse
	//  400: ResponseError
	//  404: ResponseError
	//  500: ResponseError
	ctx := r.Context()

	// Parse the linkID from the uri.
	id := mux.Vars(r)["id"]
	linkID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid link id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Send the linkID to the link manager for removal.
	err = s.Storage.Remove(ctx, assets.LinkID(linkID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewLinkFilter creates an assets links filter from the the given request's URL query parameters.
func NewLinkFilter(r *http.Request) (assets.LinkFilter, error) {
	q := r.URL.Query()
	filter := assets.LinkFilter{
		Limit: assets.DefaultLinkFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return assets.LinkFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = assets.PlayerID(ownerID)
	}

	if values := q["locationID"]; len(values) > 0 {
		locationID, err := uuid.Parse(values[0])
		if err != nil {
			return assets.LinkFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.LocationID = assets.RoomID(locationID)
	}

	if values := q["destinationID"]; len(values) > 0 {
		destinationID, err := uuid.Parse(values[0])
		if err != nil {
			return assets.LinkFilter{}, fmt.Errorf("%w: invalid destinationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.DestinationID = assets.RoomID(destinationID)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return assets.LinkFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > assets.MaxLinkFilterLimit {
			return assets.LinkFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	return filter, nil
}

// TranslateLinkRequest translates a network link request to an assets link request.
func TranslateLinkRequest(l rest.LinkRequest) (assets.LinkChange, error) {
	empty := assets.LinkChange{}

	if l.Name == "" {
		return empty, fmt.Errorf("%w: empty link name", errors.ErrBadRequest)
	}
	if len(l.Name) > assets.MaxLinkNameLen {
		return empty, fmt.Errorf("%w: link name exceeds maximum length", errors.ErrBadRequest)
	}
	if l.Description == "" {
		return empty, fmt.Errorf("%w: empty link description", errors.ErrBadRequest)
	}
	if len(l.Description) > assets.MaxLinkDescriptionLen {
		return empty, fmt.Errorf("%w: link description exceeds maximum length", errors.ErrBadRequest)
	}
	ownerID, err := uuid.Parse(l.OwnerID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid ownerID: '%s'", errors.ErrBadRequest, l.OwnerID)
	}
	locID, err := uuid.Parse(l.LocationID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid locationID: '%s', %s", errors.ErrBadRequest, l.LocationID, err)
	}
	destID, err := uuid.Parse(l.DestinationID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid destinationID: '%s', %s", errors.ErrBadRequest, l.DestinationID, err)
	}

	return assets.LinkChange{
		Name:          l.Name,
		Description:   l.Description,
		OwnerID:       assets.PlayerID(ownerID),
		LocationID:    assets.RoomID(locID),
		DestinationID: assets.RoomID(destID),
	}, nil
}

// TranslateLink translates an asset link to a network link.
func TranslateLink(l *assets.Link) rest.Link {
	return rest.Link{
		ID:            l.ID.String(),
		Name:          l.Name,
		Description:   l.Description,
		OwnerID:       l.OwnerID.String(),
		LocationID:    l.LocationID.String(),
		DestinationID: l.DestinationID.String(),
		Created:       l.Created,
		Updated:       l.Updated,
	}
}
