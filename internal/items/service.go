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

package items

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	cerrors "arcadium.dev/core/errors"
	"arcadium.dev/core/log"

	"arcadium.dev/arcade/internal/arcade"
)

const (
	route string = "/items"
)

type (
	// Service is used to manage the item assets.
	Service struct {
		db *sql.DB
		h  handler
	}
)

func New(db *sql.DB) Service {
	s := &Service{db: db}
	s.h = handler{s: s}
	return *s
}

// Register sets up the http handler for this service with the given router.
func (s Service) Register(router *mux.Router) {
	r := router.PathPrefix(route).Subrouter()
	r.HandleFunc("", s.h.list).Methods(http.MethodGet)
	r.HandleFunc("/{itemID}", s.h.get).Methods(http.MethodGet)
	r.HandleFunc("", s.h.create).Methods(http.MethodPost)
	r.HandleFunc("/{itemID}", s.h.update).Methods(http.MethodPut)
	r.HandleFunc("/{itemID}", s.h.remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (Service) Name() string {
	return "items"
}

// Shutdown is a no-op since there no long running processes for this service... yet.
func (Service) Shutdown() {}

const (
	listQuery   = `SELECT item_id, name, description, owner, location, inventory, created, updated FROM items`
	getQuery    = `SELECT item_id, name, description, owner, location, inventory, created, updated FROM items WHERE item_id = $1`
	createQuery = `INSERT INTO items (name, description, owner, location, inventory) ` +
		`VALUES ($1, $2, $3, $4, $5) ` +
		`RETURNING item_id, name, description, owner, location, inventory, created, updated`
	updateQuery = `UPDATE items SET name = $2, description = $3, owner = $4, location = $5, inventory = $6,  updated = now() ` +
		`WHERE item_id = $1 ` +
		`RETURNING item_id, name, description, owner, location, inventory, created, updated`
	removeQuery = `DELETE FROM items WHERE item_id = $1`
)

func (s Service) list(ctx context.Context) ([]arcade.Item, error) {
	failMsg := "failed to list items"

	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "list items")

	// TODO: build query based on query params.

	rows, err := s.db.QueryContext(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("msg", "failed to close rows of list query", "error", err.Error())
		}
	}()

	items := make([]arcade.Item, 0)
	for rows.Next() {
		var p item
		err := rows.Scan(
			&p.id,
			&p.name,
			&p.description,
			&p.owner,
			&p.location,
			&p.inventory,
			&p.created,
			&p.updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return items, nil
}

func (s Service) get(ctx context.Context, pid string) (arcade.Item, error) {
	failMsg := "failed to get item"

	log.LoggerFromContext(ctx).With("itemID", pid).Info("msg", "get item")

	itemID, err := uuid.Parse(pid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid item id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}

	var p item
	err = s.db.QueryRowContext(ctx, getQuery, itemID).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.location,
		&p.inventory,
		&p.created,
		&p.updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return p, nil
}

func (s Service) create(ctx context.Context, req itemRequest) (arcade.Item, error) {
	failMsg := "failed to create item"

	logger := log.LoggerFromContext(ctx).With("name", req.Name)
	logger.Info("msg", "create item")

	// Validate the input.
	if req.Name == "" {
		return nil, fmt.Errorf("%s: %w: empty item name", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Name) > maxNameLen {
		return nil, fmt.Errorf("%s: %w: item name exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	if req.Description == "" {
		return nil, fmt.Errorf("%s: %w: empty item description", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Description) > maxDescriptionLen {
		return nil, fmt.Errorf("%s: %w: item description exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	ownerID, err := uuid.Parse(req.Owner)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid owner: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Owner)
	}
	locationID, err := uuid.Parse(req.Location)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid location: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Location)
	}
	inventoryID, err := uuid.Parse(req.Inventory)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid inventory: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Inventory)
	}

	// Query the database.
	var p item
	err = s.db.QueryRowContext(ctx, createQuery,
		req.Name,
		req.Description,
		ownerID,
		locationID,
		inventoryID,
	).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.location,
		&p.inventory,
		&p.created,
		&p.updated,
	)

	var pgErr *pgconn.PgError

	// A ForeignKeyViolation means the referenced ownerID, locationID or inventoryID does not exist
	// in the items table, thus we will return an invalid argument error.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%s: %w: the given owner, location, or inventory does not exist: owner '%s', location '%s', inventory '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.Owner, req.Location, req.Inventory,
		)
	}

	// A UniqueViolation means the inserted item violated a uniqueness
	// constraint. The item record already exists in the table or the name
	// is not unique.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%s: %w: item already exists", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	logger.With("itemID", p.id).Info("msg", "created item")
	return p, nil
}

func (s Service) update(ctx context.Context, pid string, req itemRequest) (arcade.Item, error) {
	failMsg := "failed to update item"

	logger := log.LoggerFromContext(ctx).With("itemID", pid, "name", req.Name)
	logger.Info("msg", "update item")

	// Validate the input.
	itemID, err := uuid.Parse(pid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid item id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}
	if req.Name == "" {
		return nil, fmt.Errorf("%s: %w: empty item name", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Name) > maxNameLen {
		return nil, fmt.Errorf("%s: %w: item name exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	if req.Description == "" {
		return nil, fmt.Errorf("%s: %w: empty item description", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Description) > maxDescriptionLen {
		return nil, fmt.Errorf("%s: %w: item description exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	ownerID, err := uuid.Parse(req.Owner)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid owner: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Owner)
	}
	locationID, err := uuid.Parse(req.Location)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid location: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Location)
	}
	inventoryID, err := uuid.Parse(req.Inventory)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid inventory: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Inventory)
	}

	// Query the database.
	var p item
	err = s.db.QueryRowContext(ctx, updateQuery,
		itemID,
		req.Name,
		req.Description,
		ownerID,
		locationID,
		inventoryID,
	).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.location,
		&p.inventory,
		&p.created,
		&p.updated,
	)

	// Tried to update a item that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced ownerID, locationID, or inventoryID does not exist
	// in the items table, thus we will return an invalid argument error.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%s: %w: the given owner, location, or inventory does not exist: owner '%s', location '%s', inventory '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.Owner, req.Location, req.Inventory,
		)
	}

	// A UniqueViolation means the inserted item violated a uniqueness
	// constraint. The item name is not unique.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%s: %w: item name is not unique", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	return p, nil
}

func (s Service) remove(ctx context.Context, pid string) error {
	failMsg := "failed to remove item"

	log.LoggerFromContext(ctx).With("itemID", pid).Info("msg", "remove item")

	itemID, err := uuid.Parse(pid)
	if err != nil {
		return fmt.Errorf("%s: %w: invalid item id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}

	_, err = s.db.ExecContext(ctx, removeQuery, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return nil
}
