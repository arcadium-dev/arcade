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

package data // import "arcadium.dev/arcade/asset/data"

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade/asset"
)

type (
	// RoomStorage is used to manage the persistent storage of room data.
	RoomStorage struct {
		DB     *sql.DB
		Driver RoomDriver
	}

	// RoomDriver abstracts away the SQL driver specific functionality.
	RoomDriver interface {
		Driver
		ListQuery(asset.RoomFilter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		RemoveQuery() string
	}
)

// List returns a slice of rooms based on the balue of the filter.
func (r RoomStorage) List(ctx context.Context, filter asset.RoomFilter) ([]*asset.Room, error) {
	failMsg := "failed to list rooms"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msg("list rooms")

	rows, err := r.DB.Query(ctx, r.Driver.ListQuery(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Err(err).Msg("failed to close rows of list query")
		}
	}()

	rooms := make([]*asset.Room, 0)
	for rows.Next() {
		var room asset.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.OwnerID,
			&room.ParentID,
			&room.Created,
			&room.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
		}
		rooms = append(rooms, &room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return rooms, nil
}

// Get returns a room given the roomID.
func (r RoomStorage) Get(ctx context.Context, roomID asset.RoomID) (*asset.Room, error) {
	failMsg := "failed to get room"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("get room: %s", roomID)

	var room asset.Room
	err := r.DB.QueryRow(ctx, r.Driver.GetQuery(), roomID).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.OwnerID,
		&room.ParentID,
		&room.Created,
		&room.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return &room, nil
}

// Create persists a new room, returning room information including the newly created roomID.
func (r RoomStorage) Create(ctx context.Context, create asset.RoomCreate) (*asset.Room, error) {
	failMsg := "failed to create room"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("create room: %s", create.Name)

	var (
		room asset.Room
	)
	err := r.DB.QueryRow(ctx, r.Driver.CreateQuery(),
		create.Name,
		create.Description,
		create.OwnerID,
		create.ParentID,
	).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.OwnerID,
		&room.ParentID,
		&room.Created,
		&room.Updated,
	)

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if r.Driver.IsForeignKeyViolation(err) {
		return nil, fmt.Errorf(
			"%s: %w: the given ownerID or parentID does not exist: ownerID '%s', parentID '%s'",
			failMsg, errors.ErrBadRequest, create.OwnerID, create.ParentID,
		)
	}

	// A UniqueViolation means the inserted room violated a uniqueness
	// constraint. The room record already exists in the table or the name
	// is not unique.
	if r.Driver.IsUniqueViolation(err) {
		return nil, fmt.Errorf("%s: %w: room name '%s' already exists", failMsg, errors.ErrBadRequest, create.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	logger.Info().Msgf("created room, id: %s", room.ID)

	return &room, nil
}

// Update updates a room, return the updated room.
func (r RoomStorage) Update(ctx context.Context, roomID asset.RoomID, update asset.RoomUpdate) (*asset.Room, error) {
	failMsg := "failed to update room"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("update room: %s", roomID)

	var room asset.Room
	err := r.DB.QueryRow(ctx, r.Driver.UpdateQuery(),
		roomID,
		update.Name,
		update.Description,
		update.OwnerID,
		update.ParentID,
	).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.OwnerID,
		&room.ParentID,
		&room.Created,
		&room.Updated,
	)

	// Tried to update a room that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if r.Driver.IsForeignKeyViolation(err) {
		return nil, fmt.Errorf(
			"%s: %w: the given ownerID or parentID does not exist: ownerID '%s', parentID '%s'",
			failMsg, errors.ErrBadRequest, update.OwnerID, update.ParentID,
		)
	}

	// A UniqueViolation means the inserted room violated a uniqueness
	// constraint. The room name is not unique.
	if r.Driver.IsUniqueViolation(err) {
		return nil, fmt.Errorf("%s: %w: room name '%s' already exists", failMsg, errors.ErrBadRequest, update.Name)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err.Error())
	}

	return &room, nil
}

// Remove deletes the room from persistent storage.
func (r RoomStorage) Remove(ctx context.Context, roomID asset.RoomID) error {
	failMsg := "failed to remove room"

	zerolog.Ctx(ctx).Info().Msgf("remove room %s", roomID)

	_, err := r.DB.Exec(ctx, r.Driver.RemoveQuery(), roomID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return nil
}
