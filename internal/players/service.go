//  Copyright 2021-2022 arcadium.dev <info@arcadium.dev>
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
	"database/sql"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	cerrors "arcadium.dev/core/errors"
)

const (
	route string = "/players"
)

type (
	// Service reports on the health of the service as a whole.
	Service struct {
		db *sql.DB
		h  handler
	}
)

func New(db *sql.DB) *Service {
	s := &Service{db: db}
	s.h = handler{s: s}
	return s
}

// Register sets up the http handler for this service with the given router.
func (s Service) Register(router *mux.Router) {
	r := router.PathPrefix(route).Subrouter()
	r.HandleFunc("", s.h.list).Methods(http.MethodGet)
	r.HandleFunc("/{playerID}", s.h.get).Methods(http.MethodGet)
	r.HandleFunc("", s.h.create).Methods(http.MethodPost)
	r.HandleFunc("/{playerID}", s.h.update).Methods(http.MethodPut)
	r.HandleFunc("/{playerID}", s.h.remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (Service) Name() string {
	return "players"
}

// Shutdown is a no-op since there no long running processes for this service.
func (Service) Shutdown() {}

type (
	// playerRequest is the payload of a player request.
	playerRequest struct {
		PlayerID    string `json:"playerID"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	// playerResponse is used as payload data for player responses.
	playerResponseData struct {
		PlayerID    string    `json:"playerID"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Home        string    `json:"home"`
		Location    string    `json:"location"`
		Inventory   []string  `json"inventory"`
		Created     time.Time `json:"created"`
	}

	// playerResponse is used to json encoded a response with a single player.
	deploymentResponse struct {
		Data playerResponseData `json:"data"`
	}

	// playersResponse is used to json encoded a response with a multiple players.
	playersResponse struct {
		Data []playerResponseData `json:"data"`
	}

	// player is the internal representation of the data related to a player.
	player struct {
		playerID    string
		name        string
		description string
		home        string
		location    string
		inventory   []string
		created     time.Time
	}
)

const (
	// Queries
	listQuery   = `SELECT player_id, name, description, home, location FROM players`
	getQuery    = `SELECT player_id, name, description, home, location FROM players WHERE player_id = $1`
	createQuery = `INSERT INTO players (player_id, name, description, home, location) VALUES ($1, $2, $3, $4, $5)`
	updateQuery = ``
	removeQuery = `DELETE FROM players WHERE player_id = $1`
)

func (s *Service) list(ctx context.Context) ([]player, error) {
	return nil, cerrors.ErrNotImplemented
}

func (s *Service) get(ctx context.Context, playerID string) (player, error) {
	return player{}, cerrors.ErrNotImplemented
}

func (s *Service) create(ctx context.Context, p player) error {
	return cerrors.ErrNotImplemented
}

func (s *Service) update(ctx context.Context, p player) error {
	return cerrors.ErrNotImplemented
}

func (s *Service) remove(ctx context.Context, playerID string) error {
	return cerrors.ErrNotImplemented
}
