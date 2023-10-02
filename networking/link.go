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

	"arcadium.dev/arcade"
	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	V1LinksRoute string = "/v1/links"
)

type (
	// LinkService services link related network requests.
	LinksService struct {
		Manager LinkManager
	}

	// LinkManager defines the expected behavior of the link manager in the domain layer.
	LinkManager interface {
		List(ctx context.Context, filter arcade.LinksFilter) ([]*arcade.Link, error)
		Get(ctx context.Context, linkID arcade.LinkID) (*arcade.Link, error)
		Create(ctx context.Context, ingressLink arcade.IngressLink) (*arcade.Link, error)
		Update(ctx context.Context, linkID arcade.LinkID, ingressLink arcade.IngressLink) (*arcade.Link, error)
		Remove(ctx context.Context, linkID arcade.LinkID) error
	}

	// IngressLink is used to request a link be created or updated.
	//
	// swagger:parameters LinkCreate LinkUpdate
	IngressLink struct {
		// Name is the name of the link.
		// in: body
		// minimum length: 1
		// maximum length: 256
		Name string `json:"name"`

		// Description is the description of the link.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		Description string `json:"description"`

		// OwnerID is the ID of the owner of the link.
		// in: body
		// minimum length: 1
		// maximum length: 4096
		OwnerID string `json:"ownerID"`

		// LocationID is the ID of the location of the link.
		// in: body
		LocationID string `json:"locationID"`

		// DestinationID is the ID of the destination of the link.
		// in: body
		DestinationID string `json:"destinationID"`
	}

	// EgressLink returns a link.
	EgressLink struct {
		// Link returns the information about a link.
		// in: body
		Link Link `json:"link"`
	}

	// LinksResponse returns multiple links.
	EgressLinks struct {
		// Links returns the information about multiple links.
		// in: body
		Links []Link `json:"links"`
	}

	// Link holds a link's information, and is sent in a response.
	//
	// swagger:parameter
	Link struct {
		// ID is the link identifier.
		// in: body
		ID string `json:"id"`

		// Name is the link name.
		// in: body
		Name string `json:"name"`

		// Description is the link description.
		// in: body
		Description string `json:"description"`

		// OwnerID is the PlayerID of the link owner.
		// in:body
		OwnerID string `json:"ownerID"`

		// LocationID is the RoomID of the link's location.
		// in: body
		LocationID string `json:"locationID"`

		// DestinationID is the RoomID of the link's destination.
		// in: body
		DestinationID string `json:"destinationID"`

		// Created is the time of the link's creation.
		// in: body
		Created arcade.Timestamp `json:"created"`

		// Updated is the time the link was last updated.
		// in: body
		Updated arcade.Timestamp `json:"updated"`
	}
)

// Register sets up the http handler for this service with the given router.
func (s LinksService) Register(router *mux.Router) {
	r := router.PathPrefix(V1LinksRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{linkID}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{linkID}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{linkID}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (LinksService) Name() string {
	return "links"
}

// Shutdown is a no-op since there no long running processes for this service.
func (LinksService) Shutdown() {}

// List handles a request to retrieve multiple links.
func (s LinksService) List(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/links List
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
	filter, err := NewLinksFilter(r)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of links.
	aLinks, err := s.Manager.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from arcade links, to local links.
	var links []Link
	for _, aLink := range aLinks {
		links = append(links, TranslateLink(aLink))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressLinks{Links: links})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve a link.
func (s LinksService) Get(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /v1/links/{linkID} Get
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
	linkID := mux.Vars(r)["linkID"]
	aLinkID, err := uuid.Parse(linkID)
	if err != nil {
		err := fmt.Errorf("%w: invalid linkID, not a well formed uuid: '%s'", errors.ErrBadRequest, linkID)
		server.Response(ctx, w, err)
		return
	}

	// Request the link from the link manager.
	aLink, err := s.Manager.Get(ctx, arcade.LinkID(aLinkID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the link to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressLink{Link: TranslateLink(aLink)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create a link.
func (s LinksService) Create(w http.ResponseWriter, r *http.Request) {
	// swagger:route POST /v1/links
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

	var ingressLink IngressLink
	err = json.Unmarshal(body, &ingressLink)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the link request to the link manager.
	aIngressLink, err := TranslateIngressLink(ingressLink)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	aLink, err := s.Manager.Create(ctx, aIngressLink)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned link for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressLink{Link: TranslateLink(aLink)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update a link.
func (s LinksService) Update(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /v1/links/{linkID}
	//
	// Update will update link based on the linkID and the link\ request in the
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
	linkID := mux.Vars(r)["linkID"]
	u, err := uuid.Parse(linkID)
	if err != nil {
		err := fmt.Errorf("%w: invalid linkID, not a well formed uuid: '%s'", errors.ErrBadRequest, linkID)
		server.Response(ctx, w, err)
		return
	}
	aLinkID := arcade.LinkID(u)

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

	// Populate the ingress link from the body.
	var ingressLink IngressLink
	err = json.Unmarshal(body, &ingressLink)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the link request.
	aIngressLink, err := TranslateIngressLink(ingressLink)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the link to the link manager.
	aLink, err := s.Manager.Update(ctx, aLinkID, aIngressLink)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(EgressLink{Link: TranslateLink(aLink)})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove a link.
func (s LinksService) Remove(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /v1/links/{linkID} Get
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
	linkID := mux.Vars(r)["linkID"]
	aLinkID, err := uuid.Parse(linkID)
	if err != nil {
		err := fmt.Errorf("%w: invalid linkID, not a well formed uuid: '%s'", errors.ErrBadRequest, linkID)
		server.Response(ctx, w, err)
		return
	}

	// Send the linkID to the link manager for removal.
	err = s.Manager.Remove(ctx, arcade.LinkID(aLinkID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// NewLinksFilter creates a LinkFilter from the the given request's URL
// query parameters.
func NewLinksFilter(r *http.Request) (arcade.LinksFilter, error) {
	q := r.URL.Query()
	filter := arcade.LinksFilter{
		Limit: arcade.DefaultLinksFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.LinksFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = arcade.PlayerID(ownerID)
	}

	if values := q["locationID"]; len(values) > 0 {
		locationID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.LinksFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.LocationID = arcade.RoomID(locationID)
	}

	if values := q["destinationID"]; len(values) > 0 {
		destinationID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.LinksFilter{}, fmt.Errorf("%w: invalid destinationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.DestinationID = arcade.RoomID(destinationID)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return arcade.LinksFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > arcade.MaxLinksFilterLimit {
			return arcade.LinksFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	return filter, nil
}

// IngressLinktranslates the link request from the http request to an arcade.LinkRequest.
func TranslateIngressLink(l IngressLink) (arcade.IngressLink, error) {
	empty := arcade.IngressLink{}

	if l.Name == "" {
		return empty, fmt.Errorf("%w: empty link name", errors.ErrBadRequest)
	}
	if len(l.Name) > arcade.MaxLinkNameLen {
		return empty, fmt.Errorf("%w: link name exceeds maximum length", errors.ErrBadRequest)
	}
	if l.Description == "" {
		return empty, fmt.Errorf("%w: empty link description", errors.ErrBadRequest)
	}
	if len(l.Description) > arcade.MaxLinkDescriptionLen {
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

	return arcade.IngressLink{
		Name:          l.Name,
		Description:   l.Description,
		OwnerID:       arcade.PlayerID(ownerID),
		LocationID:    arcade.RoomID(locID),
		DestinationID: arcade.RoomID(destID),
	}, nil
}

// TranslateLink translates an arcade link to a local link.
func TranslateLink(l *arcade.Link) Link {
	return Link{
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
