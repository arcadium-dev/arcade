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
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	cerrors "arcadium.dev/core/errors"
	"arcadium.dev/core/log"

	"arcadium.dev/arcade/internal/arcade"
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

// Shutdown is a no-op since there no long running processes for this service... yet.
func (Service) Shutdown() {}

const (
	// Queries
	listQuery   = "SELECT player_id, name, description, home, location, created, updated FROM players"
	getQuery    = "SELECT player_id, name, description, home, location, create, updated FROM players WHERE player_id = $1"
	upsertQuery = "INSERT INTO players (player_id, name, description, home, location) " +
		"VALUES ($1, $2, $3, $4, $5) " +
		"ON CONFLICT (player_id) DO " +
		"SET name = EXCLUDED.name, description = EXCLUDED.descrption, home = EXCLUDED.home, location = EXCLUDED.location"
	removeQuery = "DELETE FROM players WHERE player_id = $1"
)

func (s *Service) list(ctx context.Context) ([]arcade.Player, error) {
	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "list players")

	// TODO: build query based on query params.

	rows, err := s.db.QueryContext(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", cerrors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("msg", "failed to close rows of list query", "error", err.Error())
		}
	}()

	players := make([]arcade.Player, 0)
	for rows.Next() {
		var p player
		err := rows.Scan(
			&p.playerID,
			&p.name,
			&p.description,
			&p.home,
			&p.location,
			&p.created,
			&p.updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to list players:  %s", cerrors.ErrInternal, err)
		}
		players = append(players, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: failed to list players: %s", cerrors.ErrInternal, err)
	}

	return players, nil
}

func (s *Service) get(ctx context.Context, pid string) (arcade.Player, error) {
	log.LoggerFromContext(ctx).Info("msg", "get player")

	playerID, err := uuid.Parse(pid)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid player id: '%s'", cerrors.ErrInvalidArgument, pid)
	}

	var p player
	err = s.db.QueryRowContext(ctx, getQuery, playerID).Scan(
		&p.playerID,
		&p.name,
		&p.description,
		&p.home,
		&p.location,
		&p.location,
		&p.updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, cerrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %s", cerrors.ErrInternal, err)
	}

	return p, nil
}

func (s *Service) create(ctx context.Context, p arcade.Player) error {
	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "create player")
	return s.upsert(ctx, p)
}

func (s *Service) update(ctx context.Context, p arcade.Player) error {
	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "update player")
	return s.upsert(ctx, p)
}

func (s *Service) upsert(ctx context.Context, p arcade.Player) error {
	// Validate the input arguments.
	playerID, err := uuid.Parse(p.PlayerID())
	if err != nil {
		return fmt.Errorf(
			"%w: invalid player id: '%s'",
			cerrors.ErrInvalidArgument, p.PlayerID(),
		)
	}
	if p.Name() == "" {
		return fmt.Errorf(
			"%w: empty player name for player '%s'",
			cerrors.ErrInvalidArgument, p.PlayerID(),
		)
	}
	if p.Description() == "" {
		return fmt.Errorf(
			"%w: empty player description for player '%s'",
			cerrors.ErrInvalidArgument, p.PlayerID(),
		)
	}
	homeID, err := uuid.Parse(p.Home())
	if err != nil {
		return fmt.Errorf(
			"%w: invalid home for player '%s': home '%s'",
			cerrors.ErrInvalidArgument, p.PlayerID(), p.Home(),
		)
	}
	locationID, err := uuid.Parse(p.Location())
	if err != nil {
		return fmt.Errorf(
			"%w: invalid location for player '%s': location '%s'",
			cerrors.ErrInvalidArgument, p.PlayerID(), p.Location(),
		)
	}

	// Upsert the player into the db.
	_, err = s.db.ExecContext(ctx, upsertQuery,
		playerID,
		p.Name(),
		p.Description(),
		homeID,
		locationID,
	)

	var pgErr *pgconn.PgError

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return fmt.Errorf(
			"%w: for player '%s, the given home or location given does not exist: home '%s', location '%s'",
			cerrors.ErrInvalidArgument, p.PlayerID(), p.Home(), p.Location(),
		)
	}

	// A UniqueViolation means the inserted player violated a uniqueness
	// constraint, and that the player record already exists in the table.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return fmt.Errorf(
			"%w: player '%s' already exists",
			cerrors.ErrAlreadyExists, p.PlayerID(),
		)
	}
	if err != nil {
		return fmt.Errorf(
			"%w: unable to create player '%s':  %s",
			cerrors.ErrInternal, p.PlayerID(), err.Error(),
		)
	}

	return nil
}

func (s *Service) remove(ctx context.Context, pid string) error {
	log.LoggerFromContext(ctx).Info("msg", "remove player")

	playerID, err := uuid.Parse(pid)
	if err != nil {
		return fmt.Errorf("%w: invalid player id: '%s'", cerrors.ErrInvalidArgument, pid)
	}

	_, err = s.db.ExecContext(ctx, removeQuery, playerID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: player '%s'", cerrors.ErrNotFound, pid)
	}
	if err != nil {
		return fmt.Errorf("%w: unable to find player '%s': %s", cerrors.ErrInternal, pid, err)
	}

	return nil
}
