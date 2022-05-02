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

package storage // import "arcadium.dev/arcade/storage"

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	cerrors "arcadium.dev/core/errors"
	"arcadium.dev/core/log"

	"arcadium.dev/arcade"
)

type (
	// Rooms is used to manage the persistent storage of room assets.
	Rooms struct {
		DB     *sql.DB
		Driver arcade.StorageDriver
	}
)

// List returns a slice of rooms based on the value of the filter.
func (p Rooms) List(ctx context.Context, filter arcade.RoomsFilter) ([]arcade.Room, error) {
	failMsg := "failed to list rooms"

	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "list rooms")

	rows, err := p.DB.QueryContext(ctx, p.Driver.RoomsListQuery(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("msg", "failed to close rows of list query", "error", err.Error())
		}
	}()

	rooms := make([]arcade.Room, 0)
	for rows.Next() {
		var room arcade.Room
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
			return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return rooms, nil
}

// Get returns a single room given the roomID.
func (p Rooms) Get(ctx context.Context, roomID string) (arcade.Room, error) {
	failMsg := "failed to get room"

	log.LoggerFromContext(ctx).With("roomID", roomID).Info("msg", "get room")

	pid, err := uuid.Parse(roomID)
	if err != nil {
		return arcade.Room{}, fmt.Errorf("%s: %w: invalid room id: '%s'", failMsg, cerrors.ErrInvalidArgument, roomID)
	}

	var room arcade.Room
	err = p.DB.QueryRowContext(ctx, p.Driver.RoomsGetQuery(), pid).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.OwnerID,
		&room.ParentID,
		&room.Created,
		&room.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return arcade.Room{}, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return arcade.Room{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return room, nil
}

// Create a room given the room request, returning the creating room.
func (p Rooms) Create(ctx context.Context, req arcade.RoomRequest) (arcade.Room, error) {
	failMsg := "failed to create room"

	logger := log.LoggerFromContext(ctx).With("name", req.Name)
	logger.Info("msg", "create room")

	ownerID, parentID, err := req.Validate()
	if err != nil {
		return arcade.Room{}, fmt.Errorf("%s: %w", failMsg, err)
	}

	var room arcade.Room
	err = p.DB.QueryRowContext(ctx, p.Driver.RoomsCreateQuery(),
		req.Name,
		req.Description,
		ownerID,
		parentID,
	).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.OwnerID,
		&room.ParentID,
		&room.Created,
		&room.Updated,
	)

	// A ForeignKeyViolation means the referenced ownerID or parentID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if p.Driver.IsForeignKeyViolation(err) {
		return arcade.Room{}, fmt.Errorf(
			"%s: %w: the given ownerID or parentID does not exist: ownerID '%s', parentID '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.OwnerID, req.ParentID,
		)
	}

	// A UniqueViolation means the inserted room violated a uniqueness
	// constraint. The room record already exists in the table or the name
	// is not unique.
	if p.Driver.IsUniqueViolation(err) {
		return arcade.Room{}, fmt.Errorf("%s: %w: room already exists", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return arcade.Room{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	logger.With("roomID", room.ID).Info("msg", "created room")
	return room, nil
}

// Update a room given the room request, returning the updated room.
func (p Rooms) Update(ctx context.Context, roomID string, req arcade.RoomRequest) (arcade.Room, error) {
	failMsg := "failed to update room"

	logger := log.LoggerFromContext(ctx).With("roomID", roomID, "name", req.Name)
	logger.Info("msg", "update room")

	pid, err := uuid.Parse(roomID)
	if err != nil {
		return arcade.Room{}, fmt.Errorf("%s: %w: invalid room id: '%s'", failMsg, cerrors.ErrInvalidArgument, roomID)
	}
	ownerID, parentID, err := req.Validate()
	if err != nil {
		return arcade.Room{}, fmt.Errorf("%s: %w", failMsg, err)
	}

	var room arcade.Room
	err = p.DB.QueryRowContext(ctx, p.Driver.RoomsUpdateQuery(),
		pid,
		req.Name,
		req.Description,
		ownerID,
		parentID,
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
		return arcade.Room{}, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced ownerID or parentID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if p.Driver.IsForeignKeyViolation(err) {
		return arcade.Room{}, fmt.Errorf(
			"%s: %w: the given ownerID or parentID does not exist: ownerID '%s', parentID '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.OwnerID, req.ParentID,
		)
	}

	// A UniqueViolation means the inserted room violated a uniqueness
	// constraint. The room name is not unique.
	if p.Driver.IsUniqueViolation(err) {
		return arcade.Room{}, fmt.Errorf("%s: %w: room name is not unique", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return arcade.Room{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	return room, nil
}

// Remove deletes the given room from persistent storage.
func (p Rooms) Remove(ctx context.Context, roomID string) error {
	failMsg := "failed to remove room"

	log.LoggerFromContext(ctx).With("roomID", roomID).Info("msg", "remove room")

	pid, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("%s: %w: invalid room id: '%s'", failMsg, cerrors.ErrInvalidArgument, roomID)
	}

	_, err = p.DB.ExecContext(ctx, p.Driver.RoomsRemoveQuery(), pid)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return nil
}
