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

package data // import "arcadium.dev/arcade/assets/data"

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade/assets"
)

type (
	// PlayerStorage is used to manage the persistent storage of player data.
	PlayerStorage struct {
		DB     *sql.DB
		Driver PlayerDriver
	}

	// PlayerDriver abstracts away the SQL driver specific functionality.
	PlayerDriver interface {
		Driver
		ListQuery(assets.PlayerFilter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		RemoveQuery() string
	}
)

// List returns a slice of players based on the balue of the filter.
func (p PlayerStorage) List(ctx context.Context, filter assets.PlayerFilter) ([]*assets.Player, error) {
	failMsg := "failed to list players"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msg("list players")

	rows, err := p.DB.Query(ctx, p.Driver.ListQuery(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Err(err).Msg("failed to close rows of list query")
		}
	}()

	players := make([]*assets.Player, 0)
	for rows.Next() {
		var player assets.Player
		err := rows.Scan(
			&player.ID,
			&player.Name,
			&player.Description,
			&player.HomeID,
			&player.LocationID,
			&player.Created,
			&player.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
		}
		players = append(players, &player)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return players, nil
}

// Get returns a player given the playerID.
func (p PlayerStorage) Get(ctx context.Context, playerID assets.PlayerID) (*assets.Player, error) {
	failMsg := "failed to get player"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("get player: %s", playerID)

	var player assets.Player
	err := p.DB.QueryRow(ctx, p.Driver.GetQuery(), playerID).Scan(
		&player.ID,
		&player.Name,
		&player.Description,
		&player.HomeID,
		&player.LocationID,
		&player.Created,
		&player.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return &player, nil
}

// Create persists a new player, returning player information including the newly created playerID.
func (p PlayerStorage) Create(ctx context.Context, create assets.PlayerCreate) (*assets.Player, error) {
	failMsg := "failed to create player"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("create player: %s", create.Name)

	var (
		player assets.Player
	)
	err := p.DB.QueryRow(ctx, p.Driver.CreateQuery(),
		create.Name,
		create.Description,
		create.HomeID,
		create.LocationID,
	).Scan(
		&player.ID,
		&player.Name,
		&player.Description,
		&player.HomeID,
		&player.LocationID,
		&player.Created,
		&player.Updated,
	)

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if p.Driver.IsForeignKeyViolation(err) {
		return nil, fmt.Errorf(
			"%s: %w: the given homeID or locationID does not exist: homeID '%s', locationID '%s'",
			failMsg, errors.ErrBadRequest, create.HomeID, create.LocationID,
		)
	}

	// A UniqueViolation means the inserted player violated a uniqueness
	// constraint. The player record already exists in the table or the name
	// is not unique.
	if p.Driver.IsUniqueViolation(err) {
		return nil, fmt.Errorf("%s: %w: player name '%s' already exists", failMsg, errors.ErrBadRequest, create.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	logger.Info().Msgf("created player, id:  %s", player.ID)

	return &player, nil
}

// Update updates a player, return the updated player.
func (p PlayerStorage) Update(ctx context.Context, playerID assets.PlayerID, update assets.PlayerUpdate) (*assets.Player, error) {
	failMsg := "failed to update player"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("update player: %s", playerID)

	var player assets.Player
	err := p.DB.QueryRow(ctx, p.Driver.UpdateQuery(),
		playerID,
		update.Name,
		update.Description,
		update.HomeID,
		update.LocationID,
	).Scan(
		&player.ID,
		&player.Name,
		&player.Description,
		&player.HomeID,
		&player.LocationID,
		&player.Created,
		&player.Updated,
	)

	// Tried to update a player that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if p.Driver.IsForeignKeyViolation(err) {
		return nil, fmt.Errorf(
			"%s: %w: the given homeID or locationID does not exist: homeID '%s', locationID '%s'",
			failMsg, errors.ErrBadRequest, update.HomeID, update.LocationID,
		)
	}

	// A UniqueViolation means the inserted player violated a uniqueness
	// constraint. The player name is not unique.
	if p.Driver.IsUniqueViolation(err) {
		return nil, fmt.Errorf("%s: %w: player name '%s' already exists", failMsg, errors.ErrBadRequest, update.Name)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err.Error())
	}

	return &player, nil
}

// Remove deletes the player from persistent storage.
func (p PlayerStorage) Remove(ctx context.Context, playerID assets.PlayerID) error {
	failMsg := "failed to remove player"

	zerolog.Ctx(ctx).Info().Msgf("remove player %s", playerID)

	_, err := p.DB.Exec(ctx, p.Driver.RemoveQuery(), playerID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return nil
}
