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
	// LinkStorage is used to manage the persistent storage of link data.
	LinkStorage struct {
		DB     *sql.DB
		Driver LinkDriver
	}

	// LinkDriver abstracts away the SQL driver specific functionality.
	LinkDriver interface {
		Driver
		ListQuery(asset.LinkFilter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		RemoveQuery() string
	}
)

// List returns a slice of links based on the balue of the filter.
func (l LinkStorage) List(ctx context.Context, filter asset.LinkFilter) ([]*asset.Link, error) {
	failMsg := "failed to list links"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msg("list links")

	rows, err := l.DB.Query(ctx, l.Driver.ListQuery(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Err(err).Msg("failed to close rows of list query")
		}
	}()

	links := make([]*asset.Link, 0)
	for rows.Next() {
		var link asset.Link
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
			return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
		}
		links = append(links, &link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return links, nil
}

// Get returns a link given the linkID.
func (l LinkStorage) Get(ctx context.Context, linkID asset.LinkID) (*asset.Link, error) {
	failMsg := "failed to get link"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("get link: %s", linkID)

	var link asset.Link
	err := l.DB.QueryRow(ctx, l.Driver.GetQuery(), linkID).Scan(
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
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return &link, nil
}

// Create persists a new link, returning link information including the newly created linkID.
func (l LinkStorage) Create(ctx context.Context, create asset.LinkCreate) (*asset.Link, error) {
	failMsg := "failed to create link"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("create link: %s", create.Name)

	var (
		link asset.Link
	)
	err := l.DB.QueryRow(ctx, l.Driver.CreateQuery(),
		create.Name,
		create.Description,
		create.OwnerID,
		create.LocationID,
		create.DestinationID,
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
	// in the rooms table, thus we will return an invalid argument error.
	if l.Driver.IsForeignKeyViolation(err) {
		return nil, fmt.Errorf(
			"%s: %w: the given ownerID, locationID or destinationID does not exist: ownerID '%s', locationID '%s', destinationID '%s'",
			failMsg, errors.ErrBadRequest, create.OwnerID, create.LocationID, create.DestinationID,
		)
	}

	// A UniqueViolation means the inserted link violated a uniqueness
	// constraint. The link record already exists in the table or the name
	// is not unique.
	if l.Driver.IsUniqueViolation(err) {
		return nil, fmt.Errorf("%s: %w: link name '%s' already exists", failMsg, errors.ErrBadRequest, create.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	logger.Info().Msgf("created link, id: %s", link.ID)

	return &link, nil
}

// Update updates a link, return the updated link.
func (l LinkStorage) Update(ctx context.Context, linkID asset.LinkID, update asset.LinkUpdate) (*asset.Link, error) {
	failMsg := "failed to update link"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("update link: %s", linkID)

	var link asset.Link
	err := l.DB.QueryRow(ctx, l.Driver.UpdateQuery(),
		linkID,
		update.Name,
		update.Description,
		update.OwnerID,
		update.LocationID,
		update.DestinationID,
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
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced ownerID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if l.Driver.IsForeignKeyViolation(err) {
		return nil, fmt.Errorf(
			"%s: %w: the given ownerID, locationID or destinationID does not exist: ownerID '%s', locationID '%s', destinationID '%s'",
			failMsg, errors.ErrBadRequest, update.OwnerID, update.LocationID, update.DestinationID,
		)
	}

	// A UniqueViolation means the inserted link violated a uniqueness
	// constraint. The link name is not unique.
	if l.Driver.IsUniqueViolation(err) {
		return nil, fmt.Errorf("%s: %w: link name '%s' already exists", failMsg, errors.ErrBadRequest, update.Name)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err.Error())
	}

	return &link, nil
}

// Remove deletes the link from persistent storage.
func (l LinkStorage) Remove(ctx context.Context, linkID asset.LinkID) error {
	failMsg := "failed to remove link"

	zerolog.Ctx(ctx).Info().Msgf("remove link %s", linkID)

	_, err := l.DB.Exec(ctx, l.Driver.RemoveQuery(), linkID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return nil
}
