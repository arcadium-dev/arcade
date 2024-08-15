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
	LinkUpdateQuery = `UPDATE links SET name = $2, description = $3, owner_id = $4, location_id = $5, destination_id = $6) ` +
		`WHERE id = $1 ` +
		`RETURNING id, name, description, owner_id, location_id, destination_id, created, updated`
	LinkRemoveQuery = `DELETE FROM links WHERE id = $1`
)

// ListQuery returns the List query string given the filter.
func (LinkDriver) ListQuery(filter asset.LinkFilter) string {
	fq := ""
	cmd := "WHERE"
	if filter.OwnerID != asset.PlayerID(uuid.Nil) {
		fq += fmt.Sprintf(" %s owner_id = '%s'", cmd, filter.OwnerID)
		cmd = "AND"
	}
	if filter.LocationID != asset.RoomID(uuid.Nil) {
		fq += fmt.Sprintf(" %s location_id = '%s'", cmd, filter.LocationID)
		cmd = "AND"
	}
	if filter.DestinationID != asset.RoomID(uuid.Nil) {
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
