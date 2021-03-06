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

package http // import "arcadium.dev/arcade/http"

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	cerrors "arcadium.dev/core/errors"
	chttp "arcadium.dev/core/http"

	"arcadium.dev/arcade"
)

const (
	ItemsRoute string = "/items"
)

type (
	// Items is used to manage the item assets.
	ItemsService struct {
		Storage arcade.ItemsStorage
	}
)

// Register sets up the http handler for this service with the given router.
func (s ItemsService) Register(router *mux.Router) {
	r := router.PathPrefix(ItemsRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{itemID}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{itemID}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{itemID}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (ItemsService) Name() string {
	return "items"
}

// Shutdown is a no-op since there no long running processes for this service... yet.
func (ItemsService) Shutdown() {}

// List handles a request to retrieve multiple items.
func (s ItemsService) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: parse query params

	// Read list of items.
	items, err := s.Storage.List(ctx, arcade.ItemsFilter{})
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.NewItemsResponse(items))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve an item.
func (s ItemsService) Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemID := params["itemID"]

	ctx := r.Context()

	item, err := s.Storage.Get(ctx, itemID)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.ItemResponse{Data: item})
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create an item.
func (s ItemsService) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid json: a json encoded body is required", cerrors.ErrInvalidArgument,
		))
		return
	}

	var req arcade.ItemRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}

	item, err := s.Storage.Create(ctx, req)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.ItemResponse{Data: item})
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update an item.
func (s ItemsService) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	itemID := params["itemID"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid json: a json encoded body is required", cerrors.ErrInvalidArgument,
		))
		return
	}

	var req arcade.ItemRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}

	item, err := s.Storage.Update(ctx, itemID, req)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.ItemResponse{Data: item})
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove an item.
func (s ItemsService) Remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	itemID := params["itemID"]

	err := s.Storage.Remove(ctx, itemID)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
