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
	// Links is used to manage the persistent storage of link assets.
	Links struct {
		DB     *sql.DB
		Driver arcade.StorageDriver
	}
)

// List returns a slice of links based on the value of the filter.
func (p Links) List(ctx context.Context, filter arcade.LinksFilter) ([]arcade.Link, error) {
	failMsg := "failed to list links"

	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "list links")

	rows, err := p.DB.QueryContext(ctx, p.Driver.LinksListQuery(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("msg", "failed to close rows of list query", "error", err.Error())
		}
	}()

	links := make([]arcade.Link, 0)
	for rows.Next() {
		var link arcade.Link
		err := rows.Scan(
			&link.ID,
			&link.Name,
			&link.Description,
			&link.OwnerID,
			&link.LocationID,
			&link.DestinationID,
			&link.Created,
			&link.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return links, nil
}

// Get returns a single link given the linkID.
func (p Links) Get(ctx context.Context, linkID string) (arcade.Link, error) {
	failMsg := "failed to get link"

	log.LoggerFromContext(ctx).With("linkID", linkID).Info("msg", "get link")

	pid, err := uuid.Parse(linkID)
	if err != nil {
		return arcade.Link{}, fmt.Errorf("%s: %w: invalid link id: '%s'", failMsg, cerrors.ErrInvalidArgument, linkID)
	}

	var link arcade.Link
	err = p.DB.QueryRowContext(ctx, p.Driver.LinksGetQuery(), pid).Scan(
		&link.ID,
		&link.Name,
		&link.Description,
		&link.OwnerID,
		&link.LocationID,
		&link.DestinationID,
		&link.Created,
		&link.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return arcade.Link{}, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return arcade.Link{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return link, nil
}

// Create a link given the link request, returning the creating link.
func (p Links) Create(ctx context.Context, req arcade.LinkRequest) (arcade.Link, error) {
	failMsg := "failed to create link"

	logger := log.LoggerFromContext(ctx).With("name", req.Name)
	logger.Info("msg", "create link")

	ownerID, locationID, destinationID, err := req.Validate()
	if err != nil {
		return arcade.Link{}, fmt.Errorf("%s: %w", failMsg, err)
	}

	var link arcade.Link
	err = p.DB.QueryRowContext(ctx, p.Driver.LinksCreateQuery(),
		req.Name,
		req.Description,
		ownerID,
		locationID,
		destinationID,
	).Scan(
		&link.ID,
		&link.Name,
		&link.Description,
		&link.OwnerID,
		&link.LocationID,
		&link.DestinationID,
		&link.Created,
		&link.Updated,
	)

	// A ForeignKeyViolation means the referenced ownerID or locationID does not exist
	// in the links table, thus we will return an invalid argument error.
	if p.Driver.IsForeignKeyViolation(err) {
		return arcade.Link{}, fmt.Errorf(
			"%s: %w: the given ownerID, locationID, or destinationID does not exist: ownerID '%s', locationID '%s', destinationID '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.OwnerID, req.LocationID, req.DestinationID,
		)
	}

	// A UniqueViolation means the inserted link violated a uniqueness
	// constraint. The link record already exists in the table or the name
	// is not unique.
	if p.Driver.IsUniqueViolation(err) {
		return arcade.Link{}, fmt.Errorf("%s: %w: link already exists", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return arcade.Link{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	logger.With("linkID", link.ID).Info("msg", "created link")
	return link, nil
}

// Update a link given the link request, returning the updated link.
func (p Links) Update(ctx context.Context, linkID string, req arcade.LinkRequest) (arcade.Link, error) {
	failMsg := "failed to update link"

	logger := log.LoggerFromContext(ctx).With("linkID", linkID, "name", req.Name)
	logger.Info("msg", "update link")

	pid, err := uuid.Parse(linkID)
	if err != nil {
		return arcade.Link{}, fmt.Errorf("%s: %w: invalid link id: '%s'", failMsg, cerrors.ErrInvalidArgument, linkID)
	}
	ownerID, locationID, destinationID, err := req.Validate()
	if err != nil {
		return arcade.Link{}, fmt.Errorf("%s: %w", failMsg, err)
	}

	var link arcade.Link
	err = p.DB.QueryRowContext(ctx, p.Driver.LinksUpdateQuery(),
		pid,
		req.Name,
		req.Description,
		ownerID,
		locationID,
		destinationID,
	).Scan(
		&link.ID,
		&link.Name,
		&link.Description,
		&link.OwnerID,
		&link.LocationID,
		&link.DestinationID,
		&link.Created,
		&link.Updated,
	)

	// Tried to update a link that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return arcade.Link{}, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced ownerID or locationID does not exist
	// in the links table, thus we will return an invalid argument error.
	if p.Driver.IsForeignKeyViolation(err) {
		return arcade.Link{}, fmt.Errorf(
			"%s: %w: the given ownerID, locationID, or destinationID does not exist: ownerID '%s', locationID '%s', destinationID '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.OwnerID, req.LocationID, req.DestinationID,
		)
	}

	// A UniqueViolation means the inserted link violated a uniqueness
	// constraint. The link name is not unique.
	if p.Driver.IsUniqueViolation(err) {
		return arcade.Link{}, fmt.Errorf("%s: %w: link name is not unique", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return arcade.Link{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	return link, nil
}

// Remove deletes the given link from persistent storage.
func (p Links) Remove(ctx context.Context, linkID string) error {
	failMsg := "failed to remove link"

	log.LoggerFromContext(ctx).With("linkID", linkID).Info("msg", "remove link")

	pid, err := uuid.Parse(linkID)
	if err != nil {
		return fmt.Errorf("%s: %w: invalid link id: '%s'", failMsg, cerrors.ErrInvalidArgument, linkID)
	}

	_, err = p.DB.ExecContext(ctx, p.Driver.LinksRemoveQuery(), pid)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return nil
}
