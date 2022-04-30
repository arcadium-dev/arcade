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

package items

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	cerrors "arcadium.dev/core/errors"
	chttp "arcadium.dev/core/http"
	"github.com/gorilla/mux"

	"arcadium.dev/arcade/internal/arcade"
)

type (
	handler struct {
		s service
	}

	service interface {
		list(ctx context.Context) ([]arcade.Item, error)
		get(ctx context.Context, itemID string) (arcade.Item, error)
		create(ctx context.Context, p itemRequest) (arcade.Item, error)
		update(ctx context.Context, itemID string, p itemRequest) (arcade.Item, error)
		remove(ctx context.Context, itemID string) error
	}
)

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: parse query params

	// Read list of items.
	p, err := h.s.list(ctx)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newItemsResponse(p))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (h handler) get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemID := params["itemID"]

	ctx := r.Context()

	p, err := h.s.get(ctx, itemID)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newItemResponse(p))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (h handler) create(w http.ResponseWriter, r *http.Request) {
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

	var req itemRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}

	p, err := h.s.create(ctx, req)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newItemResponse(p))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (h handler) update(w http.ResponseWriter, r *http.Request) {
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

	var req itemRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}

	p, err := h.s.update(ctx, itemID, req)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newItemResponse(p))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (h handler) remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	itemID := params["itemID"]

	err := h.s.remove(ctx, itemID)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
