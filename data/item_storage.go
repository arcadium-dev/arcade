//  Copyright 2022-2024 arcadium.dev <info@arcadium.dev>
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

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade/asset"
)

type (
	// ItemStorage is used to manage the persistent storage of item data.
	ItemStorage struct {
		DB     *sql.DB
		Driver ItemDriver
	}

	// ItemDriver abstracts away the SQL driver specific functionality.
	ItemDriver interface {
		Driver
		ListQuery(asset.ItemFilter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		RemoveQuery() string
	}
)

// List returns a slice of items based on the balue of the filter.
func (i ItemStorage) List(ctx context.Context, filter asset.ItemFilter) ([]*asset.Item, error) {
	failMsg := "failed to list items"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msg("list items")

	rows, err := i.DB.Query(ctx, i.Driver.ListQuery(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Err(err).Msg("failed to close rows of list query")
		}
	}()

	items := make([]*asset.Item, 0)
	var locItemID, locPlayerID, locRoomID uuid.NullUUID
	for rows.Next() {
		var item asset.Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.OwnerID,
			&locItemID,
			&locPlayerID,
			&locRoomID,
			&item.Created,
			&item.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
		}

		found := false
		if locPlayerID.Valid {
			item.LocationID = asset.PlayerID(locPlayerID.UUID)
			found = true
		}
		if locRoomID.Valid {
			if !found {
				item.LocationID = asset.RoomID(locRoomID.UUID)
				found = true
			} else {
				logger.Error().Msgf("invalid location for item: %s", item.ID)
			}
		}
		if locItemID.Valid {
			if !found {
				item.LocationID = asset.ItemID(locItemID.UUID)
			} else {
				logger.Error().Msgf("invalid location for item: %s", item.ID)
			}
		}

		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return items, nil
}

// Get returns a item given the itemID.
func (i ItemStorage) Get(ctx context.Context, itemID asset.ItemID) (*asset.Item, error) {
	failMsg := "failed to get item"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("get item: %s", itemID)

	var item asset.Item
	var locItemID, locPlayerID, locRoomID uuid.NullUUID
	err := i.DB.QueryRow(ctx, i.Driver.GetQuery(), itemID).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.OwnerID,
		&locItemID,
		&locPlayerID,
		&locRoomID,
		&item.Created,
		&item.Updated,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	case err != nil:
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	setItemID(ctx, &item, locItemID, locPlayerID, locRoomID)

	return &item, nil
}

// Create persists a new item, returning item information including the newly created itemID.
func (i ItemStorage) Create(ctx context.Context, create asset.ItemCreate) (*asset.Item, error) {
	failMsg := "failed to create item"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("create item: %s", create.Name)

	var (
		item                              asset.Item
		locItemID, locPlayerID, locRoomID uuid.NullUUID
	)
	newLocItemID, newLocPlayerID, newLocRoomID := locationID(create.LocationID.Type(), create.LocationID.ID())

	err := i.DB.QueryRow(ctx, i.Driver.CreateQuery(),
		create.Name,
		create.Description,
		create.OwnerID,
		newLocItemID,
		newLocPlayerID,
		newLocRoomID,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.OwnerID,
		&locItemID,
		&locPlayerID,
		&locRoomID,
		&item.Created,
		&item.Updated,
	)

	switch {
	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	case i.Driver.IsForeignKeyViolation(err):
		return nil, fmt.Errorf(
			"%s: %w: the given ownerID or locationID does not exist: ownerID '%s', locationID '%s (%s)'",
			failMsg, errors.ErrBadRequest, create.OwnerID, create.LocationID.ID(), create.LocationID.Type(),
		)

	// A UniqueViolation means the inserted item violated a uniqueness
	// constraint. The item record already exists in the table or the name
	// is not unique.
	case i.Driver.IsUniqueViolation(err):
		return nil, fmt.Errorf("%s: %w: item name '%s' already exists", failMsg, errors.ErrBadRequest, create.Name)

	case err != nil:
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	setItemID(ctx, &item, locItemID, locPlayerID, locRoomID)

	logger.Info().Msgf("created item, id: %s", item.ID)

	return &item, nil
}

// Update updates a item, return the updated item.
func (i ItemStorage) Update(ctx context.Context, itemID asset.ItemID, update asset.ItemUpdate) (*asset.Item, error) {
	failMsg := "failed to update item"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("update item: %s", itemID)

	var (
		item                              asset.Item
		locItemID, locPlayerID, locRoomID uuid.NullUUID
	)
	newLocItemID, newLocPlayerID, newLocRoomID := locationID(update.LocationID.Type(), update.LocationID.ID())

	err := i.DB.QueryRow(ctx, i.Driver.UpdateQuery(),
		itemID,
		update.Name,
		update.Description,
		update.OwnerID,
		newLocItemID,
		newLocPlayerID,
		newLocRoomID,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.OwnerID,
		&locItemID,
		&locPlayerID,
		&locRoomID,
		&item.Created,
		&item.Updated,
	)

	switch {
	// Tried to update a item that doesn't exist.
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)

	// A ForeignKeyViolation means the referenced homeID or locationID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	case i.Driver.IsForeignKeyViolation(err):
		return nil, fmt.Errorf(
			"%s: %w: the given ownerID or locationID does not exist: ownerID '%s', locationID '%s (%s)'",
			failMsg, errors.ErrBadRequest, update.OwnerID, update.LocationID.ID(), update.LocationID.Type(),
		)

	// A UniqueViolation means the inserted item violated a uniqueness
	// constraint. The item name is not unique.
	case i.Driver.IsUniqueViolation(err):
		return nil, fmt.Errorf("%s: %w: item name '%s' already exists", failMsg, errors.ErrBadRequest, update.Name)

	case err != nil:
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err.Error())
	}

	setItemID(ctx, &item, locItemID, locPlayerID, locRoomID)

	return &item, nil
}

// Remove deletes the item from persistent storage.
func (i ItemStorage) Remove(ctx context.Context, itemID asset.ItemID) error {
	failMsg := "failed to remove item"

	zerolog.Ctx(ctx).Info().Msgf("remove item %s", itemID)

	if _, err := i.DB.Exec(ctx, i.Driver.RemoveQuery(), itemID); err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return nil
}

func locationID(t asset.LocationType, id asset.LocationID) (uuid.NullUUID, uuid.NullUUID, uuid.NullUUID) {
	var newLocItemID, newLocPlayerID, newLocRoomID uuid.NullUUID

	switch t {
	case asset.LocationTypeItem:
		newLocItemID = uuid.NullUUID{
			UUID:  uuid.UUID(id),
			Valid: true,
		}
	case asset.LocationTypePlayer:
		newLocPlayerID = uuid.NullUUID{
			UUID:  uuid.UUID(id),
			Valid: true,
		}
	case asset.LocationTypeRoom:
		newLocRoomID = uuid.NullUUID{
			UUID:  uuid.UUID(id),
			Valid: true,
		}
	}
	return newLocItemID, newLocPlayerID, newLocRoomID
}

func setItemID(ctx context.Context, item *asset.Item, locItemID, locPlayerID, locRoomID uuid.NullUUID) {
	logger := zerolog.Ctx(ctx)

	found := false
	if locPlayerID.Valid {
		item.LocationID = asset.PlayerID(locPlayerID.UUID)
		found = true
	}
	if locRoomID.Valid {
		if !found {
			item.LocationID = asset.RoomID(locRoomID.UUID)
			found = true
		} else {
			logger.Error().Msgf("invalid location for item: %s", item.ID)
		}
	}
	if locItemID.Valid {
		if !found {
			item.LocationID = asset.ItemID(locItemID.UUID)
		} else {
			logger.Error().Msgf("invalid location for item: %s", item.ID)
		}
	}
}
