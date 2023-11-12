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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"arcadium.dev/core/errors"
	"github.com/google/uuid"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/arcade/assets/rest"
)

const (
	V1ItemsRoute string = "/v1/items"
)

// ListItems ... TODO
func (c Client) ListItems(ctx context.Context, filter assets.ItemFilter) ([]*assets.Item, error) {
	failMsg := "failed to send list items request"

	// Create the request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+V1ItemsRoute, nil)
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

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	// Handle the response.
	var itemsResp rest.ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(itemsResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	var items []*assets.Item
	for _, i := range itemsResp.Items {
		aItem, err := TranslateItem(i)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", failMsg, err)
		}
		items = append(items, aItem)
	}

	return items, errors.ErrNotImplemented
}

func (c Client) Get(context.Context, assets.ItemID) (*assets.Item, error) {
	return nil, errors.ErrNotImplemented
}

func (c Client) Create(context.Context, assets.ItemCreate) (*assets.Item, error) {
	return nil, errors.ErrNotImplemented
}

func (c Client) Update(context.Context, assets.ItemID, assets.ItemUpdate) (*assets.Item, error) {
	return nil, errors.ErrNotImplemented
}

func (c Client) Remove(context.Context, assets.ItemID) error {
	return errors.ErrNotImplemented
}

// TranslateItem translates a network item into an assets item.
func TranslateItem(i rest.Item) (*assets.Item, error) {
	id, err := uuid.Parse(i.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ownerID: '%s'", err, i.OwnerID)
	}
	ownerID, err := uuid.Parse(i.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ownerID: '%s'", err, i.OwnerID)
	}
	locID, err := uuid.Parse(i.LocationID.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid locationID.ID: '%s', %s", err, i.LocationID.ID)
	}
	t := strings.ToLower(i.LocationID.Type)
	if t != "room" && t != "player" && t != "item" {
		return nil, fmt.Errorf("%w: invalid locationID.Type: '%s'", err, i.LocationID.Type)
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
