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
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

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
