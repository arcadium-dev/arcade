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
	// Items is used to manage the persistent storage of item assets.
	Items struct {
		DB     *sql.DB
		Driver arcade.StorageDriver
	}
)

// List returns a slice of items based on the value of the filter.
func (p Items) List(ctx context.Context, filter arcade.ItemsFilter) ([]arcade.Item, error) {
	failMsg := "failed to list items"

	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "list items")

	rows, err := p.DB.QueryContext(ctx, p.Driver.ItemsListQuery(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("msg", "failed to close rows of list query", "error", err.Error())
		}
	}()

	items := make([]arcade.Item, 0)
	for rows.Next() {
		var item arcade.Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.OwnerID,
			&item.LocationID,
			&item.InventoryID,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return items, nil
}

// Get returns a single item given the itemID.
func (p Items) Get(ctx context.Context, itemID string) (arcade.Item, error) {
	failMsg := "failed to get item"

	log.LoggerFromContext(ctx).With("itemID", itemID).Info("msg", "get item")

	pid, err := uuid.Parse(itemID)
	if err != nil {
		return arcade.Item{}, fmt.Errorf("%s: %w: invalid item id: '%s'", failMsg, cerrors.ErrInvalidArgument, itemID)
	}

	var item arcade.Item
	err = p.DB.QueryRowContext(ctx, p.Driver.ItemsGetQuery(), pid).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.OwnerID,
		&item.LocationID,
		&item.InventoryID,
		&item.Created,
		&item.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return arcade.Item{}, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return arcade.Item{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return item, nil
}

// Create a item given the item request, returning the creating item.
func (p Items) Create(ctx context.Context, req arcade.ItemRequest) (arcade.Item, error) {
	failMsg := "failed to create item"

	logger := log.LoggerFromContext(ctx).With("name", req.Name)
	logger.Info("msg", "create item")

	ownerID, locationID, inventoryID, err := req.Validate()
	if err != nil {
		return arcade.Item{}, fmt.Errorf("%s: %w", failMsg, err)
	}

	var item arcade.Item
	err = p.DB.QueryRowContext(ctx, p.Driver.ItemsCreateQuery(),
		req.Name,
		req.Description,
		ownerID,
		locationID,
		inventoryID,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.OwnerID,
		&item.LocationID,
		&item.InventoryID,
		&item.Created,
		&item.Updated,
	)

	// A ForeignKeyViolation means the referenced ownerID or locationID does not exist
	// in the items table, thus we will return an invalid argument error.
	if p.Driver.IsForeignKeyViolation(err) {
		return arcade.Item{}, fmt.Errorf(
			"%s: %w: the given ownerID, locationID, or inventoryID does not exist: ownerID '%s', locationID '%s', inventoryID '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.OwnerID, req.LocationID, req.InventoryID,
		)
	}

	// A UniqueViolation means the inserted item violated a uniqueness
	// constraint. The item record already exists in the table or the name
	// is not unique.
	if p.Driver.IsUniqueViolation(err) {
		return arcade.Item{}, fmt.Errorf("%s: %w: item already exists", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return arcade.Item{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	logger.With("itemID", item.ID).Info("msg", "created item")
	return item, nil
}

// Update a item given the item request, returning the updated item.
func (p Items) Update(ctx context.Context, itemID string, req arcade.ItemRequest) (arcade.Item, error) {
	failMsg := "failed to update item"

	logger := log.LoggerFromContext(ctx).With("itemID", itemID, "name", req.Name)
	logger.Info("msg", "update item")

	pid, err := uuid.Parse(itemID)
	if err != nil {
		return arcade.Item{}, fmt.Errorf("%s: %w: invalid item id: '%s'", failMsg, cerrors.ErrInvalidArgument, itemID)
	}
	ownerID, locationID, inventoryID, err := req.Validate()
	if err != nil {
		return arcade.Item{}, fmt.Errorf("%s: %w", failMsg, err)
	}

	var item arcade.Item
	err = p.DB.QueryRowContext(ctx, p.Driver.ItemsUpdateQuery(),
		pid,
		req.Name,
		req.Description,
		ownerID,
		locationID,
		inventoryID,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.OwnerID,
		&item.LocationID,
		&item.InventoryID,
		&item.Created,
		&item.Updated,
	)

	// Tried to update a item that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return arcade.Item{}, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced ownerID or locationID does not exist
	// in the items table, thus we will return an invalid argument error.
	if p.Driver.IsForeignKeyViolation(err) {
		return arcade.Item{}, fmt.Errorf(
			"%s: %w: the given ownerID, locationID, or inventoryID does not exist: ownerID '%s', locationID '%s', inventoryID '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.OwnerID, req.LocationID, req.InventoryID,
		)
	}

	// A UniqueViolation means the inserted item violated a uniqueness
	// constraint. The item name is not unique.
	if p.Driver.IsUniqueViolation(err) {
		return arcade.Item{}, fmt.Errorf("%s: %w: item name is not unique", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return arcade.Item{}, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	return item, nil
}

// Remove deletes the given item from persistent storage.
func (p Items) Remove(ctx context.Context, itemID string) error {
	failMsg := "failed to remove item"

	log.LoggerFromContext(ctx).With("itemID", itemID).Info("msg", "remove item")

	pid, err := uuid.Parse(itemID)
	if err != nil {
		return fmt.Errorf("%s: %w: invalid item id: '%s'", failMsg, cerrors.ErrInvalidArgument, itemID)
	}

	_, err = p.DB.ExecContext(ctx, p.Driver.ItemsRemoveQuery(), pid)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return nil
}
