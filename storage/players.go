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

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	// Players is used to manage the persistent storage of player assets.
	Players struct {
		DB *sql.DB
	}
)

const (
	listQuery   = `SELECT player_id, name, description, home, location, created, updated FROM players`
	getQuery    = `SELECT player_id, name, description, home, location, created, updated FROM players WHERE player_id = $1`
	createQuery = `INSERT INTO players (name, description, home, location) ` +
		`VALUES ($1, $2, $3, $4) ` +
		`RETURNING player_id, name, description, home, location, created, updated`
	updateQuery = `UPDATE players SET name = $2, description = $3, home = $4, location = $5, updated = now() ` +
		`WHERE player_id = $1 ` +
		`RETURNING player_id, name, description, home, location, created, updated`
	removeQuery = `DELETE FROM players WHERE player_id = $1`
)

// List returns a slice of players based on the value of the filter.
func (p Players) List(ctx context.Context, filter arcade.PlayerFilter) ([]arcade.Player, error) {
	failMsg := "failed to list players"

	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "list players")

	// TODO: build query based on filter

	rows, err := p.DB.QueryContext(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("msg", "failed to close rows of list query", "error", err.Error())
		}
	}()

	players := make([]arcade.Player, 0)
	for rows.Next() {
		var player arcade.Player
		err := rows.Scan(
			&player.ID,
			&player.Name,
			&player.Description,
			&player.Home,
			&player.Location,
			&player.Created,
			&player.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return players, nil
}

// Get returns a single player given the playerID.
func (p Players) Get(ctx context.Context, playerID string) (arcade.Player, error) {
	failMsg := "failed to get player"

	log.LoggerFromContext(ctx).With("playerID", playerID).Info("msg", "get player")

	pid, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid player id: '%s'", failMsg, cerrors.ErrInvalidArgument, playerID)
	}

	var player arcade.Player
	err = p.DB.QueryRowContext(ctx, getQuery, pid).Scan(
		&player.ID,
		&player.Name,
		&player.Description,
		&player.Home,
		&player.Location,
		&player.Created,
		&player.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return player, nil
}

// Create a player given the player request, returning the creating player.
func (p Players) Create(ctx context.Context, req arcade.PlayerRequest) (arcade.Player, error) {
	failMsg := "failed to create player"

	logger := log.LoggerFromContext(ctx).With("name", req.Name)
	logger.Info("msg", "create player")

	homeID, locationID, err := req.Validate()
	if err != nil {
		return fmt.Error("%s: %w", failMsg, err)
	}

	var player arcade.Player
	err = p.DB.QueryRowContext(ctx, createQuery,
		req.Name,
		req.Description,
		homeID,
		locationID,
	).Scan(
		&player.ID,
		&player.Name,
		&player.Description,
		&player.Home,
		&player.Location,
		&player.Created,
		&player.Updated,
	)

	var pgErr *pgconn.PgError

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%s: %w: the given home or location does not exist: home '%s', location '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.Home, req.Location,
		)
	}

	// A UniqueViolation means the inserted player violated a uniqueness
	// constraint. The player record already exists in the table or the name
	// is not unique.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%s: %w: player already exists", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	logger.With("playerID", player.ID).Info("msg", "created player")
	return p, nil
}

// Update a player given the player request, returning the updated player.
func (p Players) Update(ctx context.Context, playerID string, req playerRequest) (arcade.Player, error) {
	failMsg := "failed to update player"

	logger := log.LoggerFromContext(ctx).With("playerID", playerID, "name", req.Name)
	logger.Info("msg", "update player")

	pid, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid player id: '%s'", failMsg, cerrors.ErrInvalidArgument, playerID)
	}
	homeID, locationID, err := req.Validate()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", failMsg, err)
	}

	var player arcde.Player
	err = s.db.QueryRowContext(ctx, updateQuery,
		pid,
		req.Name,
		req.Description,
		homeID,
		locationID,
	).Scan(
		&player.ID,
		&player.Name,
		&player.Description,
		&player.Home,
		&player.Location,
		&player.Created,
		&player.Updated,
	)

	// Tried to update a player that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%s: %w: the given home or location does not exist: home '%s', location '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.Home, req.Location,
		)
	}

	// A UniqueViolation means the inserted player violated a uniqueness
	// constraint. The player name is not unique.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%s: %w: player name is not unique", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	return player, nil
}

// Remove deletes the given player from persistent storage.
func (p Players) Remove(ctx context.Context, playerID string) error {
	failMsg := "failed to remove player"

	log.LoggerFromContext(ctx).With("playerID", playerID).Info("msg", "remove player")

	pid, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("%s: %w: invalid player id: '%s'", failMsg, cerrors.ErrInvalidArgument, playerID)
	}

	_, err = s.db.ExecContext(ctx, removeQuery, pid)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return nil
}
