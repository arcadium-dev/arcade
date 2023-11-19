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

package client // import "arcadium.dev/arcade/assets/rest/client"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/arcade/assets/rest"
)

const (
	V1ItemsRoute string = "/v1/items"
)

// ListItems returns a list of items for the given item filter.
func (c Client) ListItems(ctx context.Context, filter assets.ItemFilter) ([]*assets.Item, error) {
	failMsg := "failed to list items"

	// Create the request.
	url := fmt.Sprintf("%s%s", c.baseURL, V1ItemsRoute)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	// Add the filter parameters.
	q := req.URL.Query()
	if filter.OwnerID != assets.NilPlayer {
		q.Add("ownerID", filter.OwnerID.String())
	}
	if filter.LocationID != nil {
		if filter.LocationID.ID() != assets.NilLocationID {
			q.Add("locationID", filter.LocationID.ID().String())
			q.Add("locationType", filter.LocationID.Type().String())
		}
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

	return itemsResponse(resp.Body, failMsg)
}

// GetItem returns an item for the given item id.
func (c Client) GetItem(ctx context.Context, id assets.ItemID) (*assets.Item, error) {
	failMsg := "failed to get item"

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1ItemsRoute, id)
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

	return itemResponse(resp.Body, failMsg)
}

// CreateItem creates an item.
func (c Client) CreateItem(ctx context.Context, item assets.ItemCreate) (*assets.Item, error) {
	failMsg := "failed to create item"

	// Build the request body.
	change := TranslateItemChange(item.ItemChange)
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(change); err != nil {
		return nil, fmt.Errorf("%s: failed to encode request body: %w", failMsg, err)
	}

	// Create the request.
	url := fmt.Sprintf("%s%s", c.baseURL, V1ItemsRoute)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	zerolog.Ctx(ctx).Info().RawJSON("request", reqBody.Bytes()).Msg("create item")

	// Send the request
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return itemResponse(resp.Body, failMsg)
}

// UpdateItem updates the item with the given item update.
func (c Client) UpdateItem(ctx context.Context, id assets.ItemID, item assets.ItemUpdate) (*assets.Item, error) {
	failMsg := "failed to update item"

	// Build the request body.
	change := TranslateItemChange(item.ItemChange)
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(change); err != nil {
		return nil, fmt.Errorf("%s: failed to encode request body: %w", failMsg, err)
	}

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1ItemsRoute, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	zerolog.Ctx(ctx).Debug().RawJSON("request", reqBody.Bytes()).Msg("update item")

	// Send the request
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return itemResponse(resp.Body, failMsg)
}

// RemoveItem deletes an item.
func (c Client) RemoveItem(ctx context.Context, id assets.ItemID) error {
	failMsg := "failed to remove item"

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1ItemsRoute, id)
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

func itemsResponse(body io.ReadCloser, failMsg string) ([]*assets.Item, error) {
	var itemsResp rest.ItemsResponse
	if err := json.NewDecoder(body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	var aItems []*assets.Item
	for _, i := range itemsResp.Items {
		aItem, err := TranslateItem(i)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", failMsg, err)
		}
		aItems = append(aItems, aItem)
	}

	return aItems, nil
}

func itemResponse(body io.ReadCloser, failMsg string) (*assets.Item, error) {
	var itemResp rest.ItemResponse
	if err := json.NewDecoder(body).Decode(&itemResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	aItem, err := TranslateItem(itemResp.Item)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	return aItem, nil
}

// TranslateItem translates a network item into an assets item.
func TranslateItem(i rest.Item) (*assets.Item, error) {
	id, err := uuid.Parse(i.ID)
	if err != nil {
		return nil, fmt.Errorf("received invalid item ID: '%s': %w", i.ID, err)
	}
	ownerID, err := uuid.Parse(i.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("received invalid item ownerID: '%s': %w", i.OwnerID, err)
	}
	locID, err := uuid.Parse(i.LocationID.ID)
	if err != nil {
		return nil, fmt.Errorf("received invalid item locationID.ID: '%s': %w", i.LocationID.ID, err)
	}
	t := strings.ToLower(i.LocationID.Type)
	if t != "room" && t != "player" && t != "item" {
		return nil, fmt.Errorf("received invalid item locationID.Type: '%s'", i.LocationID.Type)
	}

	item := &assets.Item{
		ID:          assets.ItemID(id),
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     assets.PlayerID(ownerID),
		Created:     i.Created,
		Updated:     i.Updated,
	}

	switch t {
	case "room":
		item.LocationID = assets.RoomID(locID)
	case "player":
		item.LocationID = assets.PlayerID(locID)
	case "item":
		item.LocationID = assets.ItemID(locID)
	}

	return item, nil
}

// TranslateItemChange translates an asset item change struct to a network item request.
func TranslateItemChange(i assets.ItemChange) rest.ItemRequest {
	return rest.ItemRequest{
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     i.OwnerID.String(),
		LocationID:  rest.ItemLocationID{
			// ID:   i.LocationID.ID().String(),
			// Type: i.LocationID.Type().String(),
		},
	}
}
