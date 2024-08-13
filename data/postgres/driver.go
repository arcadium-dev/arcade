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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"

	"arcadium.dev/core/sql"
)

// Open opens a database.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("failed to open database: dsn required")
	}

	db, err := sql.Open(ctx, "pgx/v5", dsn, sql.WithReconnect(3), sql.WithTxRetries(3))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.DB.SetConnMaxLifetime(time.Minute * 3)
	db.DB.SetMaxOpenConns(20)
	db.DB.SetMaxIdleConns(20)

	return db, nil
}

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
