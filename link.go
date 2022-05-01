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

package arcade // import "arcadium.dev/arcade"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/core/errors"
)

const (
	MaxLinkNameLen        = 255
	MaxLinkDescriptionLen = 4096
)

type (
	// Link is the internal representation of the data related to a link.
	Link struct {
		ID            string    `json:"linkID"`
		Name          string    `json:"name"`
		Description   string    `json:"description"`
		OwnerID       string    `json:"ownerID"`
		LocationID    string    `json:"locationID"`
		DestinationID string    `json:"destinationID"`
		Created       time.Time `json:"created"`
		Updated       time.Time `json:"updated"`
	}

	// LinkRequest is the payload of a link create or update request.
	LinkRequest struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		OwnerID       string `json:"ownerID"`
		LocationID    string `json:"locationID"`
		DestinationID string `json:"destinationID"`
	}

	// LinkResponse is used to json encoded a single link response.
	LinkResponse struct {
		Data Link `json:"data"`
	}

	// LinksResponse is used to json encoded a multi-link response.
	LinksResponse struct {
		Data []Link `json:"data"`
	}

	// LinksFilter is used to filter results from a List.
	LinksFilter struct {
		// OwnerID filters for links owned by a given link.
		OwnerID *string

		// LocationID filters for links located in a location link (non-recursive).
		LocationID *string

		// DestinationID filters for links connected to the given destination.
		DestinationID *string

		// Restrict to a subset of the results.
		Offset int
		Limit  int
	}

	// LinkStorage represents the persistent storage of links.
	LinkStorage interface {
		// List returns a slice of links based on the value of the filter.
		List(ctx context.Context, filter LinksFilter) ([]Link, error)

		// Get returns a single link given the linkID.
		Get(ctx context.Context, linkID string) (Link, error)

		// Create a link given the link request, returning the creating link.
		Create(ctx context.Context, req LinkRequest) (Link, error)

		// Update a link given the link request, returning the updated link.
		Update(ctx context.Context, linkID string, req LinkRequest) (Link, error)

		// Remove deletes the given link from persistent storage.
		Remove(ctx context.Context, linkID string) error
	}
)

// Validate returns an error for an invalid link request. A vaild request
// will return the parsed owner and location UUIDs.
func (r LinkRequest) Validate() (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	if r.Name == "" {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty link name", errors.ErrInvalidArgument)
	}
	if len(r.Name) > MaxLinkNameLen {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: link name exceeds maximum length", errors.ErrInvalidArgument)
	}
	if r.Description == "" {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: empty link description", errors.ErrInvalidArgument)
	}
	if len(r.Description) > MaxLinkDescriptionLen {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: link description exceeds maximum length", errors.ErrInvalidArgument)
	}
	ownerID, err := uuid.Parse(r.OwnerID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid ownerID: '%s'", errors.ErrInvalidArgument, r.OwnerID)
	}
	locationID, err := uuid.Parse(r.LocationID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid locationID: '%s'", errors.ErrInvalidArgument, r.LocationID)
	}
	destinationID, err := uuid.Parse(r.DestinationID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("%w: invalid destinationID: '%s'", errors.ErrInvalidArgument, r.DestinationID)
	}
	return ownerID, locationID, destinationID, nil
}

// NewLinksResponse returns a links response given a slice of links.
func NewLinksResponse(rs []Link) LinksResponse {
	var resp LinksResponse
	for _, r := range rs {
		resp.Data = append(resp.Data, r)
	}
	return resp
}
