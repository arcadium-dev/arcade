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
	V1RoomRoute string = "/v1/room"
)

// ListRooms returns a list of rooms for the given room filter.
func (c Client) ListRooms(ctx context.Context, filter asset.RoomFilter) ([]*asset.Room, error) {
	failMsg := "failed to list rooms"

	// Create the request.
	url := fmt.Sprintf("%s%s", c.baseURL, V1RoomRoute)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	// Add the filter parameters.
	q := req.URL.Query()
	if filter.OwnerID != asset.NilPlayerID {
		q.Add("ownerID", filter.OwnerID.String())
	}
	if filter.ParentID != asset.NilRoomID {
		q.Add("parentID", filter.ParentID.String())
	}
	if filter.Offset > 0 {
		q.Add("offset", strconv.FormatUint(uint64(filter.Offset), 10))
	}
	if filter.Limit > 0 {
		if filter.Limit > asset.MaxRoomFilterLimit {
			return nil, fmt.Errorf("%s: room filter limit %d exceeds maximum %d", failMsg, filter.Limit, asset.MaxRoomFilterLimit)
		}
		q.Add("limit", strconv.FormatUint(uint64(filter.Limit), 10))
	}
	req.URL.RawQuery = q.Encode()

	// Send the request.
	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return roomsResponse(resp.Body, failMsg)
}

// GetRoom returns an room for the given room id.
func (c Client) GetRoom(ctx context.Context, id asset.RoomID) (*asset.Room, error) {
	failMsg := "failed to get room"

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1RoomRoute, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	// Send the request.
	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return roomResponse(resp.Body, failMsg)
}

// CreateRoom creates an room.
func (c Client) CreateRoom(ctx context.Context, room asset.RoomCreate) (*asset.Room, error) {
	failMsg := "failed to create room"

	// Build the request body.
	change, err := TranslateRoomChange(room.RoomChange)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(change); err != nil {
		return nil, fmt.Errorf("%s: failed to encode request body: %w", failMsg, err)
	}

	// Create the request.
	url := fmt.Sprintf("%s%s", c.baseURL, V1RoomRoute)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	zerolog.Ctx(ctx).Info().RawJSON("request", reqBody.Bytes()).Msg("create room")

	// Send the request
	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return roomResponse(resp.Body, failMsg)
}

// UpdateRoom updates the room with the given room update.
func (c Client) UpdateRoom(ctx context.Context, id asset.RoomID, room asset.RoomUpdate) (*asset.Room, error) {
	failMsg := "failed to update room"

	// Build the request body.
	change, err := TranslateRoomChange(room.RoomChange)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(change); err != nil {
		return nil, fmt.Errorf("%s: failed to encode request body: %w", failMsg, err)
	}

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1RoomRoute, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	zerolog.Ctx(ctx).Debug().RawJSON("request", reqBody.Bytes()).Msg("update room")

	// Send the request
	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return roomResponse(resp.Body, failMsg)
}

// RemoveRoom deletes an room.
func (c Client) RemoveRoom(ctx context.Context, id asset.RoomID) error {
	failMsg := "failed to remove room"

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1RoomRoute, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", failMsg, err)
	}

	// Send the request
	resp, err := c.Send(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return nil
}

func roomsResponse(body io.ReadCloser, failMsg string) ([]*asset.Room, error) {
	var roomsResp rest.RoomsResponse
	if err := json.NewDecoder(body).Decode(&roomsResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	var aRooms []*asset.Room
	for _, p := range roomsResp.Rooms {
		aRoom, err := TranslateRoom(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", failMsg, err)
		}
		aRooms = append(aRooms, aRoom)
	}

	return aRooms, nil
}

func roomResponse(body io.ReadCloser, failMsg string) (*asset.Room, error) {
	var roomResp rest.RoomResponse
	if err := json.NewDecoder(body).Decode(&roomResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	aRoom, err := TranslateRoom(roomResp.Room)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	return aRoom, nil
}

// TranslateRoom translates a network room into an asset room.
func TranslateRoom(p rest.Room) (*asset.Room, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("received invalid room ID: '%s': %w", p.ID, err)
	}
	ownerID, err := uuid.Parse(p.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("received invalid room ownerID: '%s': %w", p.OwnerID, err)
	}
	parentID, err := uuid.Parse(p.ParentID)
	if err != nil {
		return nil, fmt.Errorf("received invalid room parentID: '%s': %w", p.ParentID, err)
	}

	room := &asset.Room{
		ID:          asset.RoomID(id),
		Name:        p.Name,
		Description: p.Description,
		OwnerID:     asset.PlayerID(ownerID),
		ParentID:    asset.RoomID(parentID),
		Created:     p.Created,
		Updated:     p.Updated,
	}

	return room, nil
}

// TranslateRoomChange translates an asset room change struct to a network room request.
func TranslateRoomChange(i asset.RoomChange) (rest.RoomRequest, error) {
	emptyResp := rest.RoomRequest{}

	if i.Name == "" {
		return emptyResp, fmt.Errorf("attempted to send empty name in request")
	}
	if i.Description == "" {
		return emptyResp, fmt.Errorf("attempted to send empty description in request")
	}

	return rest.RoomRequest{
		Name:        i.Name,
		Description: i.Description,
		OwnerID:     i.OwnerID.String(),
		ParentID:    i.ParentID.String(),
	}, nil
}
