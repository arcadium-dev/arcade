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

package postgres // import "arcadium.dev/arcade/asset/data/postgres"

import (
	"fmt"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"arcadium.dev/arcade/asset"
)

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
	RoomUpdateQuery = `UPDATE rooms SET name = $2, description = $3, owner_id = $4, parent_id = $5 ` +
		`WHERE id = $1 ` +
		`RETURNING id, name, description, owner_id, parent_id, created, updated`
	RoomRemoveQuery = `DELETE FROM rooms WHERE id = $1`
)

// ListQuery returns the List query string given the filter.
func (RoomDriver) ListQuery(filter asset.RoomFilter) string {
	fq := ""
	cmd := "WHERE"
	if filter.OwnerID != asset.PlayerID(uuid.Nil) {
		fq += fmt.Sprintf(" %s owner_id = '%s'", cmd, filter.OwnerID)
		cmd = "AND"
	}
	if filter.ParentID != asset.RoomID(uuid.Nil) {
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
