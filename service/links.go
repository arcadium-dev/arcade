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

package service // import "arcadium.dev/arcade/service"

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/http/server"

	"arcadium.dev/arcade"
)

const (
	LinksRoute string = "/links"
)

type (
	// Links is used to manage the link assets.
	Links struct {
		Storage arcade.LinksStorage
	}
)

// Register sets up the http handler for this service with the given router.
func (s Links) Register(router *mux.Router) {
	r := router.PathPrefix(LinksRoute).Subrouter()
	r.HandleFunc("", s.List).Methods(http.MethodGet)
	r.HandleFunc("/{linkID}", s.Get).Methods(http.MethodGet)
	r.HandleFunc("", s.Create).Methods(http.MethodPost)
	r.HandleFunc("/{linkID}", s.Update).Methods(http.MethodPut)
	r.HandleFunc("/{linkID}", s.Remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (Links) Name() string {
	return "links"
}

// Shutdown is a no-op since there no long running processes for this service.
func (Links) Shutdown() {}

// List handles a request to retrieve multiple links.
func (s Links) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: parse query params

	// Read list of links.
	links, err := s.Storage.List(ctx, arcade.LinksFilter{})
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.NewLinksResponse(links))
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Get handles a request to retrieve a link.
func (s Links) Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	linkID := params["linkID"]

	ctx := r.Context()

	link, err := s.Storage.Get(ctx, linkID)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.LinkResponse{Data: link})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Create handles a request to create a link.
func (s Links) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request: %s", errors.ErrBadRequest, err,
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

	var req arcade.LinkRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	link, err := s.Storage.Create(ctx, req)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.LinkResponse{Data: link})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Update handles a request to update a link.
func (s Links) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	linkID := params["linkID"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to read request: %s", errors.ErrBadRequest, err,
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

	var req arcade.LinkRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", errors.ErrBadRequest, err,
		))
		return
	}

	link, err := s.Storage.Update(ctx, linkID, req)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arcade.LinkResponse{Data: link})
	if err != nil {
		server.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", errors.ErrInternal, err,
		))
		return
	}
}

// Remove handles a request to remove a link.
func (s Links) Remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	linkID := params["linkID"]

	err := s.Storage.Remove(ctx, linkID)
	if err != nil {
		server.Response(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
