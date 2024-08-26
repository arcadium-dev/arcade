//  Copyright 2024 arcadium.dev <info@arcadium.dev>
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

package server // import "arcadium.dev/arcade/user/rest/server"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade/asset"
	oapi "arcadium.dev/arcade/internal/user/server"
	"arcadium.dev/arcade/user"
)

const (
	V1UserRoute string = "/v1/user"
)

type (
	// UserService services user related network requests.
	UsersService struct {
		Storage UserStorage
	}

	// UserStorage defines the expected behavior of the user manager in the data layer.
	UserStorage interface {
		List(context.Context, user.Filter) ([]*user.User, error)
		Get(context.Context, user.ID) (*user.User, error)
		Create(context.Context, user.Create) (*user.User, error)
		Update(context.Context, user.ID, user.Update) (*user.User, error)
		AssociatePlayer(context.Context, user.ID, user.AssociatePlayer) (*user.User, error)
		Remove(context.Context, user.ID) error
	}
)

// Register sets up the http handler for this service with the given router.
func (s UsersService) Register(router *mux.Router) {
	options := oapi.GorillaServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			server.Response(r.Context(), w, err)
		},
	}
	oapi.HandlerWithOptions(s, options)
}

// Name returns the name of the service.
func (UsersService) Name() string {
	return "users"
}

// Shutdown is a no-op since there no long running processes for this service.
func (UsersService) Shutdown(context.Context) {}

// List handles a request to retrieve multiple users.
func (s UsersService) List(w http.ResponseWriter, r *http.Request, params oapi.ListParams) {
	ctx := r.Context()

	// Create a filter from the quesry parameters.
	filter, err := UserFilter(params)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Read list of users.
	uUsers, err := s.Storage.List(ctx, filter)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Translate from user users, to network users.
	users := make([]oapi.User, 0)
	for _, uUser := range uUsers {
		users = append(users, TranslateUser(uUser))
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(oapi.UsersResponse{Users: users})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode user list response, error %s", err)
		return
	}
}

// Get handles a request to retrieve an user.
func (s UsersService) Get(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	// Parse the userID from the uri.
	userID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid user id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Request the user from the user manager.
	user, err := s.Storage.Get(ctx, user.ID(userID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the user to be returned in the body of the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(oapi.UserResponse{User: TranslateUser(user)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode user get response, error %s", err)
		return
	}
}

// Create handles a request to create an user.
func (s UsersService) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse the user request from the body of the request.
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

	var createReq oapi.UserCreateRequest
	err = json.Unmarshal(body, &createReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Send the user request to the user manager.
	change, err := CreateChange(createReq)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	user, err := s.Storage.Create(ctx, user.Create{Change: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Prepare the returned user for delivery in the response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(oapi.UserResponse{User: TranslateUser(user)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode user create response, error %s", err)
		return
	}
}

// Update handles a request to update an user.
func (s UsersService) Update(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	// Grab the userID from the uri.
	userID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid user id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
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

	// Populate the network user from the body.
	var updateReq oapi.UserUpdateRequest
	err = json.Unmarshal(body, &updateReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the user request.
	change, err := UpdateChange(updateReq)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the user to the user manager.
	user, err := s.Storage.Update(ctx, user.ID(userID), user.Update{Change: change})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(oapi.UserResponse{User: TranslateUser(user)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode user update response, error %s", err)
		return
	}
}

// AssociatePlayer associates a player with the given user.
func (s UsersService) AssociatePlayer(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	// Grab the userID from the uri.
	userID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid user id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
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

	// Populate the network user from the body.
	var assocPlayerReq oapi.AssociatePlayerRequest
	err = json.Unmarshal(body, &assocPlayerReq)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	// Translate the user request.
	assocPlayer, err := AssocPlayer(assocPlayerReq)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Send the user to the user manager.
	user, err := s.Storage.AssociatePlayer(ctx, user.ID(userID), assocPlayer)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	err = json.NewEncoder(w).Encode(oapi.UserResponse{User: TranslateUser(user)})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Msgf("failed to encode user asociate player response, error %s", err)
		return
	}
}

// Remove handles a request to remove an user.
func (s UsersService) Remove(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	// Parse the userID from the uri.
	userID, err := uuid.Parse(id)
	if err != nil {
		err := fmt.Errorf("%w: invalid user id, not a well formed uuid: '%s'", errors.ErrBadRequest, id)
		server.Response(ctx, w, err)
		return
	}

	// Send the userID to the user manager for removal.
	err = s.Storage.Remove(ctx, user.ID(userID))
	if err != nil {
		server.Response(ctx, w, err)
		return
	}
}

// UserFilter creates an user users filter from the the given request's query parameters.
func UserFilter(params oapi.ListParams) (user.Filter, error) {
	filter := user.Filter{
		Limit: user.DefaultUserFilterLimit,
	}

	if params.Offset != nil {
		o := *params.Offset
		offset, err := strconv.Atoi(o)
		if err != nil || offset <= 0 {
			return user.Filter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, o)
		}
		filter.Offset = uint(offset)
	}

	if params.Limit != nil {
		l := *params.Limit
		limit, err := strconv.Atoi(l)
		if err != nil || limit <= 0 || limit > user.MaxUserFilterLimit {
			return user.Filter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, l)
		}
		filter.Limit = uint(limit)
	}

	return filter, nil
}

// CreateChange translates a user create request to an user change.
func CreateChange(r oapi.UserCreateRequest) (user.Change, error) {
	empty := user.Change{}

	if r.Login == "" {
		return empty, fmt.Errorf("%w: empty user login", errors.ErrBadRequest)
	}
	if len(r.Login) > user.MaxLoginLen {
		return empty, fmt.Errorf("%w: user login exceeds maximum length", errors.ErrBadRequest)
	}
	if r.PublicKey == "" {
		return empty, fmt.Errorf("%w: empty user ssh public key", errors.ErrBadRequest)
	}
	if len(r.PublicKey) > user.MaxPublicKeyLen {
		return empty, fmt.Errorf("%w: user ssh public key exceeds maximum length", errors.ErrBadRequest)
	}

	return user.Change{
		Login:     r.Login,
		PublicKey: []byte(r.PublicKey),
	}, nil
}

// UpdateChange translates a user create request to an user change.
func UpdateChange(r oapi.UserUpdateRequest) (user.Change, error) {
	empty := user.Change{}

	if r.Login == "" {
		return empty, fmt.Errorf("%w: empty user login", errors.ErrBadRequest)
	}
	if len(r.Login) > user.MaxLoginLen {
		return empty, fmt.Errorf("%w: user login exceeds maximum length", errors.ErrBadRequest)
	}
	if r.PublicKey == "" {
		return empty, fmt.Errorf("%w: empty user ssh public key", errors.ErrBadRequest)
	}
	if len(r.PublicKey) > user.MaxPublicKeyLen {
		return empty, fmt.Errorf("%w: user ssh public key exceeds maximum length", errors.ErrBadRequest)
	}

	return user.Change{
		Login:     r.Login,
		PublicKey: []byte(r.PublicKey),
	}, nil
}

// AssocPlayer translates a user player update request to an user change.
func AssocPlayer(r oapi.AssociatePlayerRequest) (user.AssociatePlayer, error) {
	empty := user.AssociatePlayer{}

	playerID, err := uuid.Parse(r.PlayerID)
	if err != nil {
		return empty, fmt.Errorf("%w: invalid playerID: '%s'", errors.ErrBadRequest, r.PlayerID)
	}

	return user.AssociatePlayer{
		PlayerID: asset.PlayerID(playerID),
	}, nil
}

// TranslateUser translates an user user to a network user.
func TranslateUser(u *user.User) oapi.User {
	return oapi.User{
		ID:        u.ID.String(),
		Login:     u.Login,
		PublicKey: string(u.PublicKey),
		PlayerID:  u.PlayerID.String(),
		Created:   u.Created,
		Updated:   u.Updated,
	}
}
