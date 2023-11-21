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

package client // import "arcadium.dev/arcade/asset/rest/client"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"arcadium.dev/arcade/asset"
	"arcadium.dev/arcade/asset/rest"
)

const (
	V1LinkRoute string = "/v1/link"
)

// ListLinks returns a list of links for the given link filter.
func (c Client) ListLinks(ctx context.Context, filter asset.LinkFilter) ([]*asset.Link, error) {
	failMsg := "failed to list links"

	// Create the request.
	url := fmt.Sprintf("%s%s", c.baseURL, V1LinkRoute)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	// Add the filter parameters.
	q := req.URL.Query()
	if filter.OwnerID != asset.NilPlayerID {
		q.Add("ownerID", filter.OwnerID.String())
	}
	if filter.LocationID != asset.NilRoomID {
		q.Add("ownerID", filter.OwnerID.String())
	}
	if filter.DestinationID != asset.NilRoomID {
		q.Add("ownerID", filter.OwnerID.String())
	}
	if filter.Offset > 0 {
		q.Add("offset", strconv.FormatUint(uint64(filter.Offset), 10))
	}
	if filter.Limit > 0 {
		q.Add("limit", strconv.FormatUint(uint64(filter.Limit), 10))
	}
	req.URL.RawQuery = q.Encode()

	// Send the request.
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return linksResponse(resp.Body, failMsg)
}

// GetLink returns an link for the given link id.
func (c Client) GetLink(ctx context.Context, id asset.LinkID) (*asset.Link, error) {
	failMsg := "failed to get link"

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1LinkRoute, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	// Send the request.
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return linkResponse(resp.Body, failMsg)
}

// CreateLink creates an link.
func (c Client) CreateLink(ctx context.Context, link asset.LinkCreate) (*asset.Link, error) {
	failMsg := "failed to create link"

	// Build the request body.
	change, err := TranslateLinkChange(link.LinkChange)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(change); err != nil {
		return nil, fmt.Errorf("%s: failed to encode request body: %w", failMsg, err)
	}

	// Create the request.
	url := fmt.Sprintf("%s%s", c.baseURL, V1LinkRoute)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	zerolog.Ctx(ctx).Info().RawJSON("request", reqBody.Bytes()).Msg("create link")

	// Send the request
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return linkResponse(resp.Body, failMsg)
}

// UpdateLink updates the link with the given link update.
func (c Client) UpdateLink(ctx context.Context, id asset.LinkID, link asset.LinkUpdate) (*asset.Link, error) {
	failMsg := "failed to update link"

	// Build the request body.
	change, err := TranslateLinkChange(link.LinkChange)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(change); err != nil {
		return nil, fmt.Errorf("%s: failed to encode request body: %w", failMsg, err)
	}

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1LinkRoute, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	zerolog.Ctx(ctx).Debug().RawJSON("request", reqBody.Bytes()).Msg("update link")

	// Send the request
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return linkResponse(resp.Body, failMsg)
}

// RemoveLink deletes an link.
func (c Client) RemoveLink(ctx context.Context, id asset.LinkID) error {
	failMsg := "failed to remove link"

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1LinkRoute, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", failMsg, err)
	}

	// Send the request
	resp, err := c.send(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return nil
}

func linksResponse(body io.ReadCloser, failMsg string) ([]*asset.Link, error) {
	var linksResp rest.LinksResponse
	if err := json.NewDecoder(body).Decode(&linksResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	var aLinks []*asset.Link
	for _, p := range linksResp.Links {
		aLink, err := TranslateLink(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", failMsg, err)
		}
		aLinks = append(aLinks, aLink)
	}

	return aLinks, nil
}

func linkResponse(body io.ReadCloser, failMsg string) (*asset.Link, error) {
	var linkResp rest.LinkResponse
	if err := json.NewDecoder(body).Decode(&linkResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	aLink, err := TranslateLink(linkResp.Link)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	return aLink, nil
}

// TranslateLink translates a network link into an asset link.
func TranslateLink(p rest.Link) (*asset.Link, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("received invalid link ID: '%s': %w", p.ID, err)
	}
	ownerID, err := uuid.Parse(p.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("received invalid link ownerID: '%s': %w", p.OwnerID, err)
	}
	locationID, err := uuid.Parse(p.LocationID)
	if err != nil {
		return nil, fmt.Errorf("received invalid link locationID: '%s': %w", p.LocationID, err)
	}
	destinationID, err := uuid.Parse(p.DestinationID)
	if err != nil {
		return nil, fmt.Errorf("received invalid link destinationID: '%s': %w", p.DestinationID, err)
	}

	link := &asset.Link{
		ID:            asset.LinkID(id),
		Name:          p.Name,
		Description:   p.Description,
		OwnerID:       asset.PlayerID(ownerID),
		LocationID:    asset.RoomID(locationID),
		DestinationID: asset.RoomID(destinationID),
		Created:       p.Created,
		Updated:       p.Updated,
	}

	return link, nil
}

// TranslateLinkChange translates an asset link change struct to a network link request.
func TranslateLinkChange(i asset.LinkChange) (rest.LinkRequest, error) {
	emptyResp := rest.LinkRequest{}

	if i.Name == "" {
		return emptyResp, fmt.Errorf("attempted to send empty name in request")
	}
	if i.Description == "" {
		return emptyResp, fmt.Errorf("attempted to send empty description in request")
	}

	return rest.LinkRequest{
		Name:          i.Name,
		Description:   i.Description,
		OwnerID:       i.OwnerID.String(),
		LocationID:    i.LocationID.String(),
		DestinationID: i.DestinationID.String(),
	}, nil
}
