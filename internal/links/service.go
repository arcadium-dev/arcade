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

package links

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
	route string = "/links"
)

type (
	// Service reports on the health of the service as a whole.
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
	r.HandleFunc("/{linkID}", s.h.get).Methods(http.MethodGet)
	r.HandleFunc("", s.h.create).Methods(http.MethodPost)
	r.HandleFunc("/{linkID}", s.h.update).Methods(http.MethodPut)
	r.HandleFunc("/{linkID}", s.h.remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (Service) Name() string {
	return "links"
}

// Shutdown is a no-op since there no long running processes for this service... yet.
func (Service) Shutdown() {}

const (
	listQuery   = `SELECT link_id, name, description, owner, location, destination, created, updated FROM links`
	getQuery    = `SELECT link_id, name, description, owner, location, destination, created, updated FROM links WHERE link_id = $1`
	createQuery = `INSERT INTO links (name, description, owner, location, destination) ` +
		`VALUES ($1, $2, $3, $4, $5) ` +
		`RETURNING link_id, name, description, owner, location, destination, created, updated`
	updateQuery = `UPDATE links SET name = $2, description = $3, owner = $4, location = $5, destination = $6,  updated = now() ` +
		`WHERE link_id = $1 ` +
		`RETURNING link_id, name, description, owner, location, destination, created, updated`
	removeQuery = `DELETE FROM links WHERE link_id = $1`
)

func (s Service) list(ctx context.Context) ([]arcade.Link, error) {
	failMsg := "failed to list links"

	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "list links")

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

	links := make([]arcade.Link, 0)
	for rows.Next() {
		var p link
		err := rows.Scan(
			&p.id,
			&p.name,
			&p.description,
			&p.owner,
			&p.location,
			&p.destination,
			&p.created,
			&p.updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
		}
		links = append(links, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return links, nil
}

func (s Service) get(ctx context.Context, pid string) (arcade.Link, error) {
	failMsg := "failed to get link"

	log.LoggerFromContext(ctx).With("linkID", pid).Info("msg", "get link")

	linkID, err := uuid.Parse(pid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid link id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}

	var p link
	err = s.db.QueryRowContext(ctx, getQuery, linkID).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.location,
		&p.destination,
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

func (s Service) create(ctx context.Context, req linkRequest) (arcade.Link, error) {
	failMsg := "failed to create link"

	logger := log.LoggerFromContext(ctx).With("name", req.Name)
	logger.Info("msg", "create link")

	// Validate the input.
	if req.Name == "" {
		return nil, fmt.Errorf("%s: %w: empty link name", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Name) > maxNameLen {
		return nil, fmt.Errorf("%s: %w: link name exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	if req.Description == "" {
		return nil, fmt.Errorf("%s: %w: empty link description", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Description) > maxDescriptionLen {
		return nil, fmt.Errorf("%s: %w: link description exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	ownerID, err := uuid.Parse(req.Owner)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid owner: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Owner)
	}
	locationID, err := uuid.Parse(req.Location)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid location: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Location)
	}
	destinationID, err := uuid.Parse(req.Destination)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid destination: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Destination)
	}

	// Query the database.
	var p link
	err = s.db.QueryRowContext(ctx, createQuery,
		req.Name,
		req.Description,
		ownerID,
		locationID,
		destinationID,
	).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.location,
		&p.destination,
		&p.created,
		&p.updated,
	)

	var pgErr *pgconn.PgError

	// A ForeignKeyViolation means the referenced ownerID, locationID or destinationID does not exist
	// in the links table, thus we will return an invalid argument error.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%s: %w: the given owner, location, or destination does not exist: owner '%s', location '%s', destination '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.Owner, req.Location, req.Destination,
		)
	}

	// A UniqueViolation means the inserted link violated a uniqueness
	// constraint. The link record already exists in the table or the name
	// is not unique.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%s: %w: link already exists", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	logger.With("linkID", p.id).Info("msg", "created link")
	return p, nil
}

func (s Service) update(ctx context.Context, pid string, req linkRequest) (arcade.Link, error) {
	failMsg := "failed to update link"

	logger := log.LoggerFromContext(ctx).With("linkID", pid, "name", req.Name)
	logger.Info("msg", "update link")

	// Validate the input.
	linkID, err := uuid.Parse(pid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid link id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}
	if req.Name == "" {
		return nil, fmt.Errorf("%s: %w: empty link name", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Name) > maxNameLen {
		return nil, fmt.Errorf("%s: %w: link name exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	if req.Description == "" {
		return nil, fmt.Errorf("%s: %w: empty link description", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Description) > maxDescriptionLen {
		return nil, fmt.Errorf("%s: %w: link description exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	ownerID, err := uuid.Parse(req.Owner)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid owner: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Owner)
	}
	locationID, err := uuid.Parse(req.Location)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid location: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Location)
	}
	destinationID, err := uuid.Parse(req.Destination)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid destination: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Destination)
	}

	// Query the database.
	var p link
	err = s.db.QueryRowContext(ctx, updateQuery,
		linkID,
		req.Name,
		req.Description,
		ownerID,
		locationID,
		destinationID,
	).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.location,
		&p.destination,
		&p.created,
		&p.updated,
	)

	// Tried to update a link that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced ownerID, locationID, or destinationID does not exist
	// in the links table, thus we will return an invalid argument error.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%s: %w: the given owner, location, or destination does not exist: owner '%s', location '%s', destination '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.Owner, req.Location, req.Destination,
		)
	}

	// A UniqueViolation means the inserted link violated a uniqueness
	// constraint. The link name is not unique.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%s: %w: link name is not unique", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	return p, nil
}

func (s Service) remove(ctx context.Context, pid string) error {
	failMsg := "failed to remove link"

	log.LoggerFromContext(ctx).With("linkID", pid).Info("msg", "remove link")

	linkID, err := uuid.Parse(pid)
	if err != nil {
		return fmt.Errorf("%s: %w: invalid link id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}

	_, err = s.db.ExecContext(ctx, removeQuery, linkID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return nil
}
