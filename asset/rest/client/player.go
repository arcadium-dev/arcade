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
	V1PlayerRoute string = "/v1/player"
)

// ListPlayers returns a list of players for the given player filter.
func (c Client) ListPlayers(ctx context.Context, filter asset.PlayerFilter) ([]*asset.Player, error) {
	failMsg := "failed to list players"

	// Create the request.
	url := fmt.Sprintf("%s%s", c.baseURL, V1PlayerRoute)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	// Add the filter parameters.
	q := req.URL.Query()
	if filter.LocationID != asset.NilRoomID {
		q.Add("locationID", filter.LocationID.String())
	}
	if filter.Offset > 0 {
		q.Add("offset", strconv.FormatUint(uint64(filter.Offset), 10))
	}
	if filter.Limit > 0 {
		if filter.Limit > asset.MaxPlayerFilterLimit {
			return nil, fmt.Errorf("%s: player filter limit %d exceeds maximum %d", failMsg, filter.Limit, asset.MaxPlayerFilterLimit)
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

	return playersResponse(resp.Body, failMsg)
}

// GetPlayer returns an player for the given player id.
func (c Client) GetPlayer(ctx context.Context, id asset.PlayerID) (*asset.Player, error) {
	failMsg := "failed to get player"

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1PlayerRoute, id)
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

	return playerResponse(resp.Body, failMsg)
}

// CreatePlayer creates an player.
func (c Client) CreatePlayer(ctx context.Context, player asset.PlayerCreate) (*asset.Player, error) {
	failMsg := "failed to create player"

	// Build the request body.
	change, err := TranslatePlayerChange(player.PlayerChange)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(change); err != nil {
		return nil, fmt.Errorf("%s: failed to encode request body: %w", failMsg, err)
	}

	// Create the request.
	url := fmt.Sprintf("%s%s", c.baseURL, V1PlayerRoute)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	zerolog.Ctx(ctx).Info().RawJSON("request", reqBody.Bytes()).Msg("create player")

	// Send the request
	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return playerResponse(resp.Body, failMsg)
}

// UpdatePlayer updates the player with the given player update.
func (c Client) UpdatePlayer(ctx context.Context, id asset.PlayerID, player asset.PlayerUpdate) (*asset.Player, error) {
	failMsg := "failed to update player"

	// Build the request body.
	change, err := TranslatePlayerChange(player.PlayerChange)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(change); err != nil {
		return nil, fmt.Errorf("%s: failed to encode request body: %w", failMsg, err)
	}

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1PlayerRoute, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	zerolog.Ctx(ctx).Debug().RawJSON("request", reqBody.Bytes()).Msg("update player")

	// Send the request
	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	defer resp.Body.Close()

	return playerResponse(resp.Body, failMsg)
}

// RemovePlayer deletes an player.
func (c Client) RemovePlayer(ctx context.Context, id asset.PlayerID) error {
	failMsg := "failed to remove player"

	// Create the request.
	url := fmt.Sprintf("%s%s/%s", c.baseURL, V1PlayerRoute, id)
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

func playersResponse(body io.ReadCloser, failMsg string) ([]*asset.Player, error) {
	var playersResp rest.PlayersResponse
	if err := json.NewDecoder(body).Decode(&playersResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	var aPlayers []*asset.Player
	for _, p := range playersResp.Players {
		aPlayer, err := TranslatePlayer(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", failMsg, err)
		}
		aPlayers = append(aPlayers, aPlayer)
	}

	return aPlayers, nil
}

func playerResponse(body io.ReadCloser, failMsg string) (*asset.Player, error) {
	var playerResp rest.PlayerResponse
	if err := json.NewDecoder(body).Decode(&playerResp); err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	aPlayer, err := TranslatePlayer(playerResp.Player)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}
	return aPlayer, nil
}

// TranslatePlayer translates a network player into an asset player.
func TranslatePlayer(p rest.Player) (*asset.Player, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("received invalid player ID: '%s': %w", p.ID, err)
	}
	homeID, err := uuid.Parse(p.HomeID)
	if err != nil {
		return nil, fmt.Errorf("received invalid player homeID: '%s': %w", p.HomeID, err)
	}
	locID, err := uuid.Parse(p.LocationID)
	if err != nil {
		return nil, fmt.Errorf("received invalid player locationID: '%s': %w", p.LocationID, err)
	}

	player := &asset.Player{
		ID:          asset.PlayerID(id),
		Name:        p.Name,
		Description: p.Description,
		HomeID:      asset.RoomID(homeID),
		LocationID:  asset.RoomID(locID),
		Created:     p.Created,
		Updated:     p.Updated,
	}

	return player, nil
}

// TranslatePlayerChange translates an asset player change struct to a network player request.
func TranslatePlayerChange(i asset.PlayerChange) (rest.PlayerRequest, error) {
	emptyResp := rest.PlayerRequest{}

	if i.Name == "" {
		return emptyResp, fmt.Errorf("attempted to send empty name in request")
	}
	if i.Description == "" {
		return emptyResp, fmt.Errorf("attempted to send empty description in request")
	}

	return rest.PlayerRequest{
		Name:        i.Name,
		Description: i.Description,
		HomeID:      i.HomeID.String(),
		LocationID:  i.LocationID.String(),
	}, nil
}
