// Copyright 2024 arcadium.dev <info@arcadium.dev>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client // import "arcadium.dev/arcade/user/client"

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"arcadium.dev/core/errors"

	"arcadium.dev/arcade/asset"
	oapi "arcadium.dev/arcade/internal/user/client"
	"arcadium.dev/arcade/user"
)

const (
	defaultTimeout = 10 * time.Second
)

var (
	ErrClient = errors.New("users client api failed")
)

type (
	// UsersClient provides a client for the users api.
	UsersClient struct {
		httpClient http.Client
		oapiClient *oapi.ClientWithResponses
	}
)

// New returns a new client for the users api.
func New(baseURL string, opts ...Option) (*UsersClient, error) {
	c := &UsersClient{
		httpClient: http.Client{
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	var err error
	c.oapiClient, err = oapi.NewClientWithResponses(
		baseURL,
		oapi.WithHTTPClient(&c.httpClient),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// List returns a list of users for the given item filter.
func (c UsersClient) List(ctx context.Context, filter user.Filter) ([]*user.User, error) {
	failMsg := "list users failed"

	resp, err := c.oapiClient.ListWithResponse(ctx, convertFilter(filter))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrClient, failMsg, err)
	}
	switch {
	case resp.JSON400 != nil:
		return nil, convertErrorResponse(resp.JSON400)
	case resp.JSON500 != nil:
		return nil, convertErrorResponse(resp.JSON500)
	case resp.JSON200 == nil:
		if resp.HTTPResponse != nil {
			return nil, fmt.Errorf("%s: unknown response, status: %s", failMsg, resp.HTTPResponse.Status)
		}
		return nil, fmt.Errorf("%s: unknown response", failMsg)
	}

	return convertUsers(resp.JSON200.Users)
}

// Get returns a user for the give user id.
func (c UsersClient) Get(ctx context.Context, id user.ID) (*user.User, error) {
	failMsg := "get user failed"

	resp, err := c.oapiClient.GetWithResponse(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrClient, failMsg, err)
	}
	switch {
	case resp.JSON400 != nil:
		return nil, convertErrorResponse(resp.JSON400)
	case resp.JSON404 != nil:
		return nil, convertErrorResponse(resp.JSON404)
	case resp.JSON500 != nil:
		return nil, convertErrorResponse(resp.JSON500)
	case resp.JSON200 == nil:
		if resp.HTTPResponse != nil {
			return nil, fmt.Errorf("%s: unknown response, status: %s", failMsg, resp.HTTPResponse.Status)
		}
		return nil, fmt.Errorf("%s: unknown response", failMsg)
	}

	return convertUser(resp.JSON200.User)
}

// Create create a new user.
func (c UsersClient) Create(ctx context.Context, req user.Create) (*user.User, error) {
	failMsg := "create user failed"

	resp, err := c.oapiClient.CreateWithResponse(ctx, convertCreate(req))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrClient, failMsg, err)
	}
	switch {
	case resp.JSON400 != nil:
		return nil, convertErrorResponse(resp.JSON400)
	case resp.JSON409 != nil:
		return nil, convertErrorResponse(resp.JSON409)
	case resp.JSON500 != nil:
		return nil, convertErrorResponse(resp.JSON500)
	case resp.JSON201 == nil:
		if resp.HTTPResponse != nil {
			return nil, fmt.Errorf("%s: unknown response, status: %s", failMsg, resp.HTTPResponse.Status)
		}
		return nil, fmt.Errorf("%s: unknown response", failMsg)
	}

	return convertUser(resp.JSON201.User)
}

// Update updates an existing user.
func (c UsersClient) Update(ctx context.Context, id user.ID, req user.Update) (*user.User, error) {
	failMsg := "update user failed"

	resp, err := c.oapiClient.UpdateWithResponse(ctx, id.String(), convertUpdate(req))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrClient, failMsg, err)
	}
	switch {
	case resp.JSON400 != nil:
		return nil, convertErrorResponse(resp.JSON400)
	case resp.JSON404 != nil:
		return nil, convertErrorResponse(resp.JSON404)
	case resp.JSON500 != nil:
		return nil, convertErrorResponse(resp.JSON500)
	case resp.JSON200 == nil:
		if resp.HTTPResponse != nil {
			return nil, fmt.Errorf("%s: unknown response, status: %s", failMsg, resp.HTTPResponse.Status)
		}
		return nil, fmt.Errorf("%s: unknown response", failMsg)
	}

	return convertUser(resp.JSON200.User)
}

// AssociatePlayer associates a player with the given user.
func (c UsersClient) AssociatePlayer(ctx context.Context, id user.ID, req user.AssociatePlayer) (*user.User, error) {
	failMsg := "associate player with user failed"

	resp, err := c.oapiClient.AssociatePlayerWithResponse(ctx, id.String(), convertAssociatePlayer(req))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrClient, failMsg, err)
	}
	switch {
	case resp.JSON400 != nil:
		return nil, convertErrorResponse(resp.JSON400)
	case resp.JSON404 != nil:
		return nil, convertErrorResponse(resp.JSON404)
	case resp.JSON500 != nil:
		return nil, convertErrorResponse(resp.JSON500)
	case resp.JSON200 == nil:
		if resp.HTTPResponse != nil {
			return nil, fmt.Errorf("%s: unknown response, status: %s", failMsg, resp.HTTPResponse.Status)
		}
		return nil, fmt.Errorf("%s: unknown response", failMsg)
	}

	return convertUser(resp.JSON200.User)
}

// Remove delete the given user.
func (c UsersClient) Remove(ctx context.Context, id user.ID) error {
	failMsg := "remove user failed"

	resp, err := c.oapiClient.RemoveWithResponse(ctx, id.String())
	if err != nil {
		return err
	}
	switch {
	case resp.JSON500 != nil:
		return convertErrorResponse(resp.JSON500)
	case resp.HTTPResponse != nil:
		if resp.HTTPResponse.StatusCode != http.StatusOK {
			return fmt.Errorf("%s: unknown response, status: %s", failMsg, resp.HTTPResponse.Status)
		}
	}

	return nil
}

// Helpers

func convertFilter(filter user.Filter) *oapi.ListParams {
	params := &oapi.ListParams{}
	if filter.Offset > 0 {
		offset := strconv.Itoa(int(filter.Offset))
		params.Offset = &offset
	}
	if filter.Limit > 0 {
		l := int(filter.Limit)
		if l > user.MaxUserFilterLimit {
			l = user.MaxUserFilterLimit
		}
		limit := strconv.Itoa(l)
		params.Limit = &limit
	}
	return params
}

func convertUsers(users []oapi.User) ([]*user.User, error) {
	if len(users) <= 0 {
		return nil, nil
	}

	us := make([]*user.User, len(users))
	for i, user := range users {
		u, err := convertUser(user)
		if err != nil {
			return nil, err
		}
		us[i] = u
	}
	return us, nil
}

func convertUser(u oapi.User) (*user.User, error) {
	id, err := uuid.Parse(u.ID)
	if err != nil {
		return nil, fmt.Errorf("%w, %w: received invalid user ID: '%s': %w", ErrClient, errors.ErrBadRequest, u.ID, err)
	}
	playerID, err := uuid.Parse(u.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("%w, %w: received invalid user playerID: '%s': %w", ErrClient, errors.ErrBadRequest, u.PlayerID, err)
	}

	return &user.User{
		ID:        user.ID(id),
		Login:     u.Login,
		PublicKey: []byte(u.PublicKey),
		PlayerID:  asset.PlayerID(playerID),
		Created:   u.Created,
		Updated:   u.Updated,
	}, nil
}

func convertCreate(req user.Create) oapi.UserCreateRequest {
	return oapi.UserCreateRequest{
		Login:     req.Login,
		PublicKey: string(req.PublicKey),
	}
}

func convertUpdate(req user.Update) oapi.UserUpdateRequest {
	return oapi.UserUpdateRequest{
		Login:     req.Login,
		PublicKey: string(req.PublicKey),
	}
}

func convertAssociatePlayer(req user.AssociatePlayer) oapi.AssociatePlayerRequest {
	return oapi.AssociatePlayerRequest{
		PlayerID: req.PlayerID.String(),
	}
}

func convertErrorResponse(resp *oapi.ErrorResponse) error {
	errMap := map[int]errors.HTTPError{
		http.StatusBadRequest: errors.ErrBadRequest,
		http.StatusNotFound:   errors.ErrNotFound,
		http.StatusConflict:   errors.ErrConflict,
	}
	err := errors.ErrInternal
	if e, ok := errMap[resp.Status]; ok {
		err = e
	}
	return fmt.Errorf("%w: error from users server '%s'", err, resp.Detail)
}
