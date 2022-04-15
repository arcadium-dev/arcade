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
	listQuery   = `SELECT player_id, name, description, home, location, created, updated FROM players`
	getQuery    = `SELECT player_id, name, description, home, location, created, updated FROM players WHERE player_id = $1`
	createQuery = `INSERT INTO players (name, description, home, location) ` +
		`VALUES ($1, $2, $3, $4) ` +
		`RETURNING player_id, name, home, location, created, updated`
	updateQuery = `UPDATE players SET name = $2, description = $3, home = $4, location = $5 ` +
		`WHERE player_id = $1` +
		`RETURNING player_id, name, home, location, created, updated`
	removeQuery = `DELETE FROM players WHERE player_id = $1`
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
			&p.id,
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
		&p.id,
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

func (s *Service) create(ctx context.Context, req playerRequest) (arcade.Player, error) {
	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "create player")

	// Validate the input.
	if req.Name == "" {
		return nil, fmt.Errorf("%w: empty player name", cerrors.ErrInvalidArgument)
	}
	if len(req.Name) > maxNameLen {
		return nil, fmt.Errorf("%w: player name exceeds maximum length", cerrors.ErrInvalidArgument)
	}
	if req.Description == "" {
		return nil, fmt.Errorf("%w: empty player description", cerrors.ErrInvalidArgument)
	}
	if len(req.Description) > maxDescriptionLen {
		return nil, fmt.Errorf("%w: player description exceeds maximum length ", cerrors.ErrInvalidArgument)
	}
	homeID, err := uuid.Parse(req.Home)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: invalid home: '%s'", cerrors.ErrInvalidArgument, req.Home)
	}
	locationID, err := uuid.Parse(req.Location)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid location: '%s'", cerrors.ErrInvalidArgument, req.Location)
	}

	// Query the database.
	var p player
	err = s.db.QueryRowContext(ctx, createQuery,
		req.Name,
		req.Description,
		homeID,
		locationID,
	).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.home,
		&p.location,
		&p.created,
		&p.updated,
	)

	var pgErr *pgconn.PgError

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%w: the given home or location given does not exist: home '%s', location '%s'",
			cerrors.ErrInvalidArgument, req.Home, req.Location,
		)
	}

	// A UniqueViolation means the inserted player violated a uniqueness
	// constraint, and that the player record already exists in the table.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%w: player already exists", cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: unable to create player:  %s", cerrors.ErrInternal, err.Error())
	}
	return p, nil
}

func (s *Service) update(ctx context.Context, pid string, req playerRequest) (arcade.Player, error) {
	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "update player")

	// Validate the input.
	playerID, err := uuid.Parse(pid)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid player id: '%s'", cerrors.ErrInvalidArgument, pid)
	}
	if req.Name == "" {
		return nil, fmt.Errorf("%w: empty player name", cerrors.ErrInvalidArgument)
	}
	if len(req.Name) > maxNameLen {
		return nil, fmt.Errorf("%w: player name exceeds maximum length", cerrors.ErrInvalidArgument)
	}
	if req.Description == "" {
		return nil, fmt.Errorf("%w: empty player description", cerrors.ErrInvalidArgument)
	}
	if len(req.Description) > maxDescriptionLen {
		return nil, fmt.Errorf("%w: player description exceeds maximum length ", cerrors.ErrInvalidArgument)
	}
	homeID, err := uuid.Parse(req.Home)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid home: '%s'", cerrors.ErrInvalidArgument, req.Home)
	}
	locationID, err := uuid.Parse(req.Location)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid location: '%s'", cerrors.ErrInvalidArgument, req.Location)
	}

	// Query the database.
	var p player
	err = s.db.QueryRowContext(ctx, updateQuery,
		playerID,
		req.Name,
		req.Description,
		homeID,
		locationID,
	).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.home,
		&p.location,
		&p.created,
		&p.updated,
	)

	// Tried to update a player that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, cerrors.ErrNotFound
	}

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%w: the given home or location given does not exist: home '%s', location '%s'",
			cerrors.ErrInvalidArgument, req.Home, req.Location,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: unable to update player '%s':  %s", cerrors.ErrInternal, pid, err.Error())
	}

	return p, nil
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
