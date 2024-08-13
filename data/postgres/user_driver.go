//  Copyright 2024 arcadium.dev <info@arcadium.dev>
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

package postgres // import "arcadium.dev/arcade/user/data/postgres"

import (
	"arcadium.dev/arcade/user"
)

type (
	UserDriver struct {
		Driver
	}
)

const (
	UserListQuery   = `SELECT id, name, public_key, player_id, created, updated FROM users`
	UserGetQuery    = `SELECT id, name, public_key, player_id, created, updated FROM users WHERE id = $1`
	UserCreateQuery = `INSERT INTO users (name, public_key, player_id) ` +
		`VALUES ($1, $2, $3 ) ` +
		`RETURNING id, name, public_key, player_id, created, updated`
	UserUpdateQuery = `UPDATE users SET name = $2, public_key = $3, player_id = $4 ` +
		`WHERE id = $1 ` +
		`RETURNING id, name, public_key, player_id, created, updated`
	UserRemoveQuery = `DELETE FROM users WHERE id = $1`
)

// ListQuery returns the List query string given the filter.
func (UserDriver) ListQuery(filter user.Filter) string {
	fq := limitAndOffset(filter.Limit, filter.Offset)
	return UserListQuery + fq
}

// GetQuery returns the Get query string.
func (UserDriver) GetQuery() string { return UserGetQuery }

// CreateQuery returns the Create query string.
func (UserDriver) CreateQuery() string { return UserCreateQuery }

// UpdateQuery returns the Update query string.
func (UserDriver) UpdateQuery() string { return UserUpdateQuery }

// RemoveQuery returns the Remove query string.
func (UserDriver) RemoveQuery() string { return UserRemoveQuery }
