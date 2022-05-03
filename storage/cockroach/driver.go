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

package cockroach // import "arcadium.dev/cockroach"

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"arcadium.dev/arcade"
)

const (
	// player queries

	playersListQuery   = `SELECT player_id, name, description, home_id, location_id, created, updated FROM players`
	playersGetQuery    = `SELECT player_id, name, description, home_id, location_id, created, updated FROM players WHERE player_id = $1`
	playersCreateQuery = `INSERT INTO players (name, description, home_id, location_id) ` +
		`VALUES ($1, $2, $3, $4) ` +
		`RETURNING player_id, name, description, home_id, location_id, created, updated`
	playersUpdateQuery = `UPDATE players SET name = $2, description = $3, home_id = $4, location_id = $5, updated = now() ` +
		`WHERE player_id = $1 ` +
		`RETURNING player_id, name, description, home_id, location_id, created, updated`
	playersRemoveQuery = `DELETE FROM players WHERE player_id = $1`

	// room queries

	roomsListQuery   = `SELECT room_id, name, description, owner_id, parent_id, created, updated FROM rooms`
	roomsGetQuery    = `SELECT room_id, name, description, owner_id, parent_id, created, updated FROM rooms WHERE room_id = $1`
	roomsCreateQuery = `INSERT INTO rooms (name, description, owner_id, parent_id) ` +
		`VALUES ($1, $2, $3, $4) ` +
		`RETURNING room_id, name, description, owner_id, parent_id, created, updated`
	roomsUpdateQuery = `UPDATE rooms SET name = $2, description = $3, owner_id = $4, parent_id = $5, updated = now() ` +
		`WHERE room_id = $1 ` +
		`RETURNING room_id, name, description, owner_id, parent_id, created, updated`
	roomsRemoveQuery = `DELETE FROM rooms WHERE room_id = $1`

	// link queries

	linksListQuery   = `SELECT link_id, name, description, owner_id, location_id, destination_id, created, updated FROM links`
	linksGetQuery    = `SELECT link_id, name, description, owner_id, location_id, destination_id, created, updated FROM links WHERE link_id = $1`
	linksCreateQuery = `INSERT INTO links (name, description, owner_id, location_id, destination_id) ` +
		`VALUES ($1, $2, $3, $4, $5) ` +
		`RETURNING link_id, name, description, owner_id, location_id, destination_id, created, updated`
	linksUpdateQuery = `UPDATE links SET name = $2, description = $3, owner_id = $4, location_id = $5, destination_id = $6,  updated = now() ` +
		`WHERE link_id = $1 ` +
		`RETURNING link_id, name, description, owner_id, location_id, destination_id, created, updated`
	linksRemoveQuery = `DELETE FROM links WHERE link_id = $1`

	// item queries

	itemsListQuery   = `SELECT item_id, name, description, owner_id, location_id, inventory_id, created, updated FROM items`
	itemsGetQuery    = `SELECT item_id, name, description, owner_id, location_id, inventory_id, created, updated FROM items WHERE item_id = $1`
	itemsCreateQuery = `INSERT INTO items (name, description, owner_id, location_id, inventory_id) ` +
		`VALUES ($1, $2, $3, $4, $5) ` +
		`RETURNING item_id, name, description, owner_id, location_id, inventory_id, created, updated`
	itemsUpdateQuery = `UPDATE items SET name = $2, description = $3, owner_id = $4, location_id = $5, inventory_id = $6,  updated = now() ` +
		`WHERE item_id = $1 ` +
		`RETURNING item_id, name, description, owner_id, location_id, inventory_id, created, updated`
	itemsRemoveQuery = `DELETE FROM items WHERE item_id = $1`
)

type (
	Driver struct{}
)

// PlayersListQuery returns the List query string given the filter.
func (Driver) PlayersListQuery(arcade.PlayersFilter) string {
	return playersListQuery
}

// PlayersGetQuery returns the Get query string.
func (Driver) PlayersGetQuery() string {
	return playersGetQuery
}

// PlayersCreateQuery returns the Create query string.
func (Driver) PlayersCreateQuery() string {
	return playersCreateQuery
}

// PlayersUpdateQuery returns the update query string.
func (Driver) PlayersUpdateQuery() string {
	return playersUpdateQuery
}

// PlayersRemoveQuery returns the Remove query string.
func (Driver) PlayersRemoveQuery() string {
	return playersRemoveQuery
}

// RoomListQuery returns the List query string given the filter.
func (Driver) RoomsListQuery(arcade.RoomsFilter) string {
	return roomsListQuery
}

// RoomsGetQuery returns the Get query string.
func (Driver) RoomsGetQuery() string {
	return roomsGetQuery
}

// RoomsCreateQuery returns the Create query string.
func (Driver) RoomsCreateQuery() string {
	return roomsCreateQuery
}

// RoomsUpdateQuery returns the Update query string.
func (Driver) RoomsUpdateQuery() string {
	return roomsUpdateQuery
}

// RoomsRemoveQuery returns the Remove query string.
func (Driver) RoomsRemoveQuery() string {
	return roomsRemoveQuery
}

// LinksListQuery returns the List query string given the filter.
func (Driver) LinksListQuery(arcade.LinksFilter) string {
	return linksListQuery
}

// LinksGetQuery returns the Get query string.
func (Driver) LinksGetQuery() string {
	return linksGetQuery
}

// LinksCreateQuery returns the Create query string.
func (Driver) LinksCreateQuery() string {
	return linksCreateQuery
}

// LinksUpdateQuery returns the Update query string.
func (Driver) LinksUpdateQuery() string {
	return linksUpdateQuery
}

// LinksRemoveQuery returns the Remove query string.
func (Driver) LinksRemoveQuery() string {
	return linksRemoveQuery
}

// ItemsListQuery returns the List query string given the filter.
func (Driver) ItemsListQuery(arcade.ItemsFilter) string {
	return itemsListQuery
}

// ItemsGetQuery returns the Get query string.
func (Driver) ItemsGetQuery() string {
	return itemsGetQuery
}

// ItemsCreateQuery returns the Create query string.
func (Driver) ItemsCreateQuery() string {
	return itemsCreateQuery
}

// ItemsUpdateQuery returns the Update query string.
func (Driver) ItemsUpdateQuery() string {
	return itemsUpdateQuery
}

// ItemsRemoveQuery returns the Remove query string.
func (Driver) ItemsRemoveQuery() string {
	return itemsRemoveQuery
}

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
