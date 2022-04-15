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

package players

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	cerrors "arcadium.dev/core/errors"
	chttp "arcadium.dev/core/http"
	"arcadium.dev/core/log"
	"github.com/gorilla/mux"

	"arcadium.dev/arcade/internal/arcade"
)

type (
	handler struct {
		s service
	}

	service interface {
		list(ctx context.Context) ([]arcade.Player, error)
		get(ctx context.Context, playerID string) (arcade.Player, error)
		create(ctx context.Context, p playerRequest) (arcade.Player, error)
		update(ctx context.Context, playerID string, p playerRequest) (arcade.Player, error)
		remove(ctx context.Context, playerID string) error
	}
)

func (h handler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: parse query params

	// Read list of players.
	p, err := h.s.list(ctx)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	// Return list as body.
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newPlayersResponse(p))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to create response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (h handler) get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	playerID := params["playerID"]

	ctx := r.Context()
	logger := log.LoggerFromContext(ctx).With("playerID", playerID)
	ctx = log.NewContextWithLogger(ctx, logger)

	p, err := h.s.get(ctx, playerID)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newPlayerResponse(p))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (h handler) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.LoggerFromContext(ctx)

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

	var req playerRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}

	logger = logger.With("name", req.Name)
	ctx = log.NewContextWithLogger(ctx, logger)

	p, err := h.s.create(ctx, req)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newPlayerResponse(p))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (h handler) update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	playerID := params["playerID"]

	ctx := r.Context()
	logger := log.LoggerFromContext(ctx).With("playerID", playerID)
	ctx = log.NewContextWithLogger(ctx, logger)

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

	var req playerRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: invalid body: %s", cerrors.ErrInvalidArgument, err,
		))
		return
	}

	p, err := h.s.update(ctx, playerID, req)
	if err != nil {
		chttp.Response(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newPlayerResponse(p))
	if err != nil {
		chttp.Response(ctx, w, fmt.Errorf(
			"%w: unable to write response: %s", cerrors.ErrInternal, err,
		))
		return
	}
}

func (h handler) remove(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	playerID := params["playerID"]

	ctx := r.Context()
	logger := log.LoggerFromContext(ctx).With("playerID", playerID)
	ctx = log.NewContextWithLogger(ctx, logger)

	err := h.s.remove(ctx, playerID)
	if err != nil {
		chttp.Response(r.Context(), w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
