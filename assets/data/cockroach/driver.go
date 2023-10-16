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

package cockroach // import "arcadium.dev/arcade/assets/data/cockroach"

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"arcadium.dev/core/sql"

	"arcadium.dev/arcade/assets"
)

// Open opens a database.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("failed to open database: dsn required")
	}

	db, err := sql.Open(ctx, "pgx", dsn, sql.WithReconnect(3), sql.WithTxRetries(3))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.DB.SetConnMaxLifetime(time.Minute * 3)
	db.DB.SetMaxOpenConns(20)
	db.DB.SetMaxIdleConns(20)

	return db, nil
}

type (
	ItemDriver struct {
		Driver
	}
)

const (
	ItemListQuery   = `SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items`
	ItemGetQuery    = `SELECT id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated FROM items WHERE id = $1`
	ItemCreateQuery = `INSERT INTO items (name, description, owner_id, location_item_id, location_player_id, location_room_id) ` +
		`VALUES ($1, $2, $3, $4, $5, $6) ` +
		`RETURNING id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated`
	ItemUpdateQuery = `UPDATE items SET name = $2, description = $3, owner_id = $4, location_item_id = $5, location_player_id = $6, location_room_id = $7, updated = now() ` +
		`WHERE id = $1 ` +
		`RETURNING id, name, description, owner_id, location_item_id, location_player_id, location_room_id, created, updated`
	ItemRemoveQuery = `DELETE FROM items WHERE id = $1`
)

// ListQuery returns the List query string given the filter.
func (ItemDriver) ListQuery(filter assets.ItemFilter) string {
	fq := ""
	cmd := "WHERE"
	if filter.OwnerID != assets.PlayerID(uuid.Nil) {
		fq += fmt.Sprintf(" %s owner_id = '%s'", cmd, filter.OwnerID)
		cmd = "AND"
	}
	if filter.LocationID != nil {
		locationID := filter.LocationID.ID()
		if locationID != assets.LocationID(uuid.Nil) {
			switch filter.LocationID.Type() {
			case assets.LocationTypeItem:
				fq += fmt.Sprintf(" %s location_item_id = '%s'", cmd, locationID)
			case assets.LocationTypePlayer:
				fq += fmt.Sprintf(" %s location_player_id = '%s'", cmd, locationID)
			case assets.LocationTypeRoom:
				fq += fmt.Sprintf(" %s location_room_id = '%s'", cmd, locationID)
			}
		}
	}
	fq += limitAndOffset(filter.Limit, filter.Offset)
	return ItemListQuery + fq
}

// GetQuery returns the Get query string.
func (ItemDriver) GetQuery() string { return ItemGetQuery }

// CreateQuery returns the Create query string.
func (ItemDriver) CreateQuery() string { return ItemCreateQuery }

// UpdateQuery returns the Update query string.
func (ItemDriver) UpdateQuery() string { return ItemUpdateQuery }

// RemoveQuery returns the Remove query string.
func (ItemDriver) RemoveQuery() string { return ItemRemoveQuery }

type (
	LinkDriver struct {
		Driver
	}
)

const (
	LinkListQuery   = `SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links`
	LinkGetQuery    = `SELECT id, name, description, owner_id, location_id, destination_id, created, updated FROM links WHERE id = $1`
	LinkCreateQuery = `INSERT INTO links (name, description, owner_id, location_id, destination_id) ` +
		`VALUES ($1, $2, $3, $4, $5) ` +
		`RETURNING id, name, description, owner_id, location_id, destination_id, created, updated`
	LinkUpdateQuery = `UPDATE links SET name = $2, description = $3, owner_id = $4, location_id = $5, destination_id = $6,  updated = now() ` +
		`WHERE id = $1 ` +
		`RETURNING id, name, description, owner_id, location_id, destination_id, created, updated`
	LinkRemoveQuery = `DELETE FROM links WHERE id = $1`
)

// ListQuery returns the List query string given the filter.
func (LinkDriver) ListQuery(filter assets.LinkFilter) string {
	fq := ""
	cmd := "WHERE"
	if filter.OwnerID != assets.PlayerID(uuid.Nil) {
		fq += fmt.Sprintf(" %s owner_id = '%s'", cmd, filter.OwnerID)
		cmd = "AND"
	}
	if filter.LocationID != assets.RoomID(uuid.Nil) {
		fq += fmt.Sprintf(" %s location_id = '%s'", cmd, filter.LocationID)
		cmd = "AND"
	}
	if filter.DestinationID != assets.RoomID(uuid.Nil) {
		fq += fmt.Sprintf(" %s destination_id = '%s'", cmd, filter.DestinationID)
	}
	fq += limitAndOffset(filter.Limit, filter.Offset)
	return LinkListQuery + fq
}

// GetQuery returns the Get query string.
func (LinkDriver) GetQuery() string { return LinkGetQuery }

// CreateQuery returns the Create query string.
func (LinkDriver) CreateQuery() string { return LinkCreateQuery }

// UpdateQuery returns the Update query string.
func (LinkDriver) UpdateQuery() string { return LinkUpdateQuery }

// RemoveQuery returns the Remove query string.
func (LinkDriver) RemoveQuery() string { return LinkRemoveQuery }

type (
	PlayerDriver struct {
		Driver
	}
)

const (
	PlayerListQuery   = `SELECT id, name, description, home_id, location_id, created, updated FROM players`
	PlayerGetQuery    = `SELECT id, name, description, home_id, location_id, created, updated FROM players WHERE id = $1`
	PlayerCreateQuery = `INSERT INTO players (name, description, home_id, location_id) ` +
		`VALUES ($1, $2, $3, $4) ` +
		`RETURNING id, name, description, home_id, location_id, created, updated`
	PlayerUpdateQuery = `UPDATE players SET name = $2, description = $3, home_id = $4, location_id = $5, updated = now() ` +
		`WHERE id = $1 ` +
		`RETURNING id, name, description, home_id, location_id, created, updated`
	PlayerRemoveQuery = `DELETE FROM players WHERE id = $1`
)

// ListQuery returns the List query string given the filter.
func (PlayerDriver) ListQuery(filter assets.PlayerFilter) string {
	fq := ""
	if filter.LocationID != assets.RoomID(uuid.Nil) {
		fq += fmt.Sprintf(" WHERE location_id = '%s'", filter.LocationID)
	}
	fq += limitAndOffset(filter.Limit, filter.Offset)
	return PlayerListQuery + fq
}

// GetQuery returns the Get query string.
func (PlayerDriver) GetQuery() string { return PlayerGetQuery }

// CreateQuery returns the Create query string.
func (PlayerDriver) CreateQuery() string { return PlayerCreateQuery }

// UpdateQuery returns the update query string.
func (PlayerDriver) UpdateQuery() string { return PlayerUpdateQuery }

// RemoveQuery returns the Remove query string.
func (PlayerDriver) RemoveQuery() string { return PlayerRemoveQuery }

type (
	RoomDriver struct {
		Driver
	}
)

const (
	RoomListQuery   = `SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms`
	RoomGetQuery    = `SELECT id, name, description, owner_id, parent_id, created, updated FROM rooms WHERE id = $1`
	RoomCreateQuery = `INSERT INTO rooms (name, description, owner_id, parent_id) ` +
		`VALUES ($1, $2, $3, $4) ` +
		`RETURNING id, name, description, owner_id, parent_id, created, updated`
	RoomUpdateQuery = `UPDATE rooms SET name = $2, description = $3, owner_id = $4, parent_id = $5, updated = now() ` +
		`WHERE id = $1 ` +
		`RETURNING id, name, description, owner_id, parent_id, created, updated`
	RoomRemoveQuery = `DELETE FROM rooms WHERE id = $1`
)

// ListQuery returns the List query string given the filter.
func (RoomDriver) ListQuery(filter assets.RoomFilter) string {
	fq := ""
	cmd := "WHERE"
	if filter.OwnerID != assets.PlayerID(uuid.Nil) {
		fq += fmt.Sprintf(" %s owner_id = '%s'", cmd, filter.OwnerID)
		cmd = "AND"
	}
	if filter.ParentID != assets.RoomID(uuid.Nil) {
		fq += fmt.Sprintf(" %s parent_id = '%s'", cmd, filter.ParentID)
	}
	fq += limitAndOffset(filter.Limit, filter.Offset)
	return RoomListQuery + fq
}

// GetQuery returns the Get query string.
func (RoomDriver) GetQuery() string { return RoomGetQuery }

// CreateQuery returns the Create query string.
func (RoomDriver) CreateQuery() string { return RoomCreateQuery }

// UpdateQuery returns the Update query string.
func (RoomDriver) UpdateQuery() string { return RoomUpdateQuery }

// RoomsRemoveQuery returns the Remove query string.
func (RoomDriver) RemoveQuery() string { return RoomRemoveQuery }

type (
	Driver struct{}
)

// IsForeignKeyViolation returns true if the given error is a foreign key violation error.
func (Driver) IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return true
	}
	return false
}

// IsUniqueViolation returns true if the given error is a unique violation error.
func (Driver) IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return true
	}
	return false
}

func limitAndOffset(limit, offset uint) string {
	fq := ""
	if limit > 0 {
		fq += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		fq += fmt.Sprintf(" OFFSET %d", offset)
	}
	return fq
}
