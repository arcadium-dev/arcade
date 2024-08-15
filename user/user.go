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

package user // import "arcadium.dev/arcade/user"

import (
	"database/sql/driver"

	"github.com/google/uuid"

	"arcadium.dev/arcade"
	"arcadium.dev/arcade/asset"
)

const (
	MaxLoginLen     = 256
	MaxPublicKeyLen = 4096

	DefaultUserFilterLimit = 50
	MaxUserFilterLimit     = 100
)

type (
	// UserID is the unique identifier of an user.
	ID uuid.UUID
)

func (i ID) String() string               { return uuid.UUID(i).String() }
func (i *ID) Scan(src any) error          { return (*uuid.UUID)(i).Scan(src) }
func (i ID) Value() (driver.Value, error) { return uuid.UUID(i).Value() }

type (
	// User is the internal representation of an user.
	User struct {
		ID        ID
		Login     string
		PublicKey []byte
		PlayerID  asset.PlayerID
		Created   arcade.Timestamp
		Updated   arcade.Timestamp
	}

	// Filter is used to filter results from a list of all users.
	Filter struct {
		// Offset is used to restrict to a subset of the results,
		// indicating the initial offset into the set of results.
		Offset uint

		// Limit is used to restrict to a subset of the results,
		// indicating the maximum number of results to return.
		Limit uint
	}

	// Create is used to create a user.
	Create struct {
		Change
	}

	// Update is used to update an user.
	Update struct {
		Change
	}

	// UserChange holds information to change an user.
	Change struct {
		Login     string
		PublicKey []byte
		PlayerID  asset.PlayerID
	}
)
