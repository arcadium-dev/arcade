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
	"fmt"

	"github.com/google/uuid"

	"arcadium.dev/arcade/assets"
)

const (
	/*
		// Item Queries

		ItemsListQuery   = `SELECT item_id, name, description, owner_id, location_id, inventory_id, created, updated FROM items`
		ItemsGetQuery    = `SELECT item_id, name, description, owner_id, location_id, inventory_id, created, updated FROM items WHERE item_id = $1`
		ItemsCreateQuery = `INSERT INTO items (name, description, owner_id, location_id, inventory_id) ` +
			`VALUES ($1, $2, $3, $4, $5) ` +
			`RETURNING item_id, name, description, owner_id, location_id, inventory_id, created, updated`
		ItemsUpdateQuery = `UPDATE items SET name = $2, description = $3, owner_id = $4, location_id = $5, inventory_id = $6,  updated = now() ` +
			`WHERE item_id = $1 ` +
			`RETURNING item_id, name, description, owner_id, location_id, inventory_id, created, updated`
		ItemsRemoveQuery = `DELETE FROM items WHERE item_id = $1`

		// Link Queries

		LinksListQuery   = `SELECT link_id, name, description, owner_id, location_id, destination_id, created, updated FROM links`
		LinksGetQuery    = `SELECT link_id, name, description, owner_id, location_id, destination_id, created, updated FROM links WHERE link_id = $1`
		LinksCreateQuery = `INSERT INTO links (name, description, owner_id, location_id, destination_id) ` +
			`VALUES ($1, $2, $3, $4, $5) ` +
			`RETURNING link_id, name, description, owner_id, location_id, destination_id, created, updated`
		LinksUpdateQuery = `UPDATE links SET name = $2, description = $3, owner_id = $4, location_id = $5, destination_id = $6,  updated = now() ` +
			`WHERE link_id = $1 ` +
			`RETURNING link_id, name, description, owner_id, location_id, destination_id, created, updated`
		LinksRemoveQuery = `DELETE FROM links WHERE link_id = $1`
	*/

	// Player Queries

	PlayersListQuery   = `SELECT id, name, description, home_id, location_id, created, updated FROM players`
	PlayersGetQuery    = `SELECT id, name, description, home_id, location_id, created, updated FROM players WHERE player_id = $1`
	PlayersCreateQuery = `INSERT INTO players (name, description, home_id, location_id) ` +
		`VALUES ($1, $2, $3, $4) ` +
		`RETURNING player_id, name, description, home_id, location_id, created, updated`
	PlayersUpdateQuery = `UPDATE players SET name = $2, description = $3, home_id = $4, location_id = $5, updated = now() ` +
		`WHERE player_id = $1 ` +
		`RETURNING player_id, name, description, home_id, location_id, created, updated`
	PlayersRemoveQuery = `DELETE FROM players WHERE player_id = $1`

	/*
		// Room Queries

		RoomsListQuery   = `SELECT room_id, name, description, owner_id, parent_id, created, updated FROM rooms`
		RoomsGetQuery    = `SELECT room_id, name, description, owner_id, parent_id, created, updated FROM rooms WHERE room_id = $1`
		RoomsCreateQuery = `INSERT INTO rooms (name, description, owner_id, parent_id) ` +
			`VALUES ($1, $2, $3, $4) ` +
			`RETURNING room_id, name, description, owner_id, parent_id, created, updated`
		RoomsUpdateQuery = `UPDATE rooms SET name = $2, description = $3, owner_id = $4, parent_id = $5, updated = now() ` +
			`WHERE room_id = $1 ` +
			`RETURNING room_id, name, description, owner_id, parent_id, created, updated`
		RoomsRemoveQuery = `DELETE FROM rooms WHERE room_id = $1`
	*/
)

type (
	Driver struct{}
)

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

/*
// ItemsListQuery returns the List query string given the filter.
func (Driver) ItemsListQuery(arcade.ItemsFilter) string {
	return ItemsListQuery
}

// ItemsGetQuery returns the Get query string.
func (Driver) ItemsGetQuery() string {
	return ItemsGetQuery
}

// ItemsCreateQuery returns the Create query string.
func (Driver) ItemsCreateQuery() string {
	return ItemsCreateQuery
}

// ItemsUpdateQuery returns the Update query string.
func (Driver) ItemsUpdateQuery() string {
	return ItemsUpdateQuery
}

// ItemsRemoveQuery returns the Remove query string.
func (Driver) ItemsRemoveQuery() string {
	return ItemsRemoveQuery
}
*/

/*
// LinksListQuery returns the List query string given the filter.
func (Driver) LinksListQuery(arcade.LinksFilter) string {
	return LinksListQuery
}

// LinksGetQuery returns the Get query string.
func (Driver) LinksGetQuery() string {
	return LinksGetQuery
}

// LinksCreateQuery returns the Create query string.
func (Driver) LinksCreateQuery() string {
	return LinksCreateQuery
}

// LinksUpdateQuery returns the Update query string.
func (Driver) LinksUpdateQuery() string {
	return LinksUpdateQuery
}

// LinksRemoveQuery returns the Remove query string.
func (Driver) LinksRemoveQuery() string {
	return LinksRemoveQuery
}
*/

// PlayersListQuery returns the List query string given the filter.
func (Driver) PlayersListQuery(filter assets.PlayersFilter) string {
	fq := ""
	if filter.LocationID != assets.RoomID(uuid.Nil) {
		fq += fmt.Sprintf(" WHERE location_id = '%s'", filter.LocationID)
	}
	fq += limitAndOffset(filter.Limit, filter.Offset)
	return PlayersListQuery + fq
}

// PlayersGetQuery returns the Get query string.
func (Driver) PlayersGetQuery() string {
	return PlayersGetQuery
}

// PlayersCreateQuery returns the Create query string.
func (Driver) PlayersCreateQuery() string {
	return PlayersCreateQuery
}

// PlayersUpdateQuery returns the update query string.
func (Driver) PlayersUpdateQuery() string {
	return PlayersUpdateQuery
}

// PlayersRemoveQuery returns the Remove query string.
func (Driver) PlayersRemoveQuery() string {
	return PlayersRemoveQuery
}

/*
// RoomListQuery returns the List query string given the filter.
func (Driver) RoomsListQuery(arcade.RoomsFilter) string {
	return RoomsListQuery
}

// RoomsGetQuery returns the Get query string.
func (Driver) RoomsGetQuery() string {
	return RoomsGetQuery
}

// RoomsCreateQuery returns the Create query string.
func (Driver) RoomsCreateQuery() string {
	return RoomsCreateQuery
}

// RoomsUpdateQuery returns the Update query string.
func (Driver) RoomsUpdateQuery() string {
	return RoomsUpdateQuery
}

// RoomsRemoveQuery returns the Remove query string.
func (Driver) RoomsRemoveQuery() string {
	return RoomsRemoveQuery
}
*/

/*
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
*/
