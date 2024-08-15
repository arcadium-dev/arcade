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

	"arcadium.dev/arcade/asset"
)

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
	PlayerUpdateQuery = `UPDATE players SET name = $2, description = $3, home_id = $4, location_id = $5 ` +
		`WHERE id = $1 ` +
		`RETURNING id, name, description, home_id, location_id, created, updated`
	PlayerRemoveQuery = `DELETE FROM players WHERE id = $1`
)

// ListQuery returns the List query string given the filter.
func (PlayerDriver) ListQuery(filter asset.PlayerFilter) string {
	fq := ""
	if filter.LocationID != asset.RoomID(uuid.Nil) {
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
