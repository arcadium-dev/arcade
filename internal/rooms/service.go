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

package rooms

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
	route string = "/rooms"
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
	r.HandleFunc("/{roomID}", s.h.get).Methods(http.MethodGet)
	r.HandleFunc("", s.h.create).Methods(http.MethodPost)
	r.HandleFunc("/{roomID}", s.h.update).Methods(http.MethodPut)
	r.HandleFunc("/{roomID}", s.h.remove).Methods(http.MethodDelete)
}

// Name returns the name of the service.
func (Service) Name() string {
	return "rooms"
}

// Shutdown is a no-op since there no long running processes for this service... yet.
func (Service) Shutdown() {}

const (
	listQuery   = `SELECT room_id, name, description, owner, parent, created, updated FROM rooms`
	getQuery    = `SELECT room_id, name, description, owner, parent, created, updated FROM rooms WHERE room_id = $1`
	createQuery = `INSERT INTO rooms (name, description, owner, parent) ` +
		`VALUES ($1, $2, $3, $4) ` +
		`RETURNING room_id, name, description, owner, parent, created, updated`
	updateQuery = `UPDATE rooms SET name = $2, description = $3, owner = $4, parent = $5, updated = now() ` +
		`WHERE room_id = $1 ` +
		`RETURNING room_id, name, description, owner, parent, created, updated`
	removeQuery = `DELETE FROM rooms WHERE room_id = $1`
)

func (s Service) list(ctx context.Context) ([]arcade.Room, error) {
	failMsg := "failed to list rooms"

	logger := log.LoggerFromContext(ctx)
	logger.Info("msg", "list rooms")

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

	rooms := make([]arcade.Room, 0)
	for rows.Next() {
		var p room
		err := rows.Scan(
			&p.id,
			&p.name,
			&p.description,
			&p.owner,
			&p.parent,
			&p.created,
			&p.updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
		}
		rooms = append(rooms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return rooms, nil
}

func (s Service) get(ctx context.Context, pid string) (arcade.Room, error) {
	failMsg := "failed to get room"

	log.LoggerFromContext(ctx).With("roomID", pid).Info("msg", "get room")

	roomID, err := uuid.Parse(pid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid room id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}

	var p room
	err = s.db.QueryRowContext(ctx, getQuery, roomID).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.parent,
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

func (s Service) create(ctx context.Context, req roomRequest) (arcade.Room, error) {
	failMsg := "failed to create room"

	logger := log.LoggerFromContext(ctx).With("name", req.Name)
	logger.Info("msg", "create room")

	// Validate the input.
	if req.Name == "" {
		return nil, fmt.Errorf("%s: %w: empty room name", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Name) > maxNameLen {
		return nil, fmt.Errorf("%s: %w: room name exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	if req.Description == "" {
		return nil, fmt.Errorf("%s: %w: empty room description", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Description) > maxDescriptionLen {
		return nil, fmt.Errorf("%s: %w: room description exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	ownerID, err := uuid.Parse(req.Owner)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid owner: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Owner)
	}
	parentID, err := uuid.Parse(req.Parent)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid parent: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Parent)
	}

	// Query the database.
	var p room
	err = s.db.QueryRowContext(ctx, createQuery,
		req.Name,
		req.Description,
		ownerID,
		parentID,
	).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.parent,
		&p.created,
		&p.updated,
	)

	var pgErr *pgconn.PgError

	// A ForeignKeyViolation means the referenced ownerID or parentID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%s: %w: the given owner or parent given does not exist: owner '%s', parent '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.Owner, req.Parent,
		)
	}

	// A UniqueViolation means the inserted room violated a uniqueness
	// constraint. The room record already exists in the table or the name
	// is not unique.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%s: %w: room already exists", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	logger.With("roomID", p.id).Info("msg", "created room")
	return p, nil
}

func (s Service) update(ctx context.Context, pid string, req roomRequest) (arcade.Room, error) {
	failMsg := "failed to update room"

	logger := log.LoggerFromContext(ctx).With("roomID", pid, "name", req.Name)
	logger.Info("msg", "update room")

	// Validate the input.
	roomID, err := uuid.Parse(pid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid room id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}
	if req.Name == "" {
		return nil, fmt.Errorf("%s: %w: empty room name", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Name) > maxNameLen {
		return nil, fmt.Errorf("%s: %w: room name exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	if req.Description == "" {
		return nil, fmt.Errorf("%s: %w: empty room description", failMsg, cerrors.ErrInvalidArgument)
	}
	if len(req.Description) > maxDescriptionLen {
		return nil, fmt.Errorf("%s: %w: room description exceeds maximum length", failMsg, cerrors.ErrInvalidArgument)
	}
	ownerID, err := uuid.Parse(req.Owner)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid owner: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Owner)
	}
	parentID, err := uuid.Parse(req.Parent)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: invalid parent: '%s'", failMsg, cerrors.ErrInvalidArgument, req.Parent)
	}

	// Query the database.
	var p room
	err = s.db.QueryRowContext(ctx, updateQuery,
		roomID,
		req.Name,
		req.Description,
		ownerID,
		parentID,
	).Scan(
		&p.id,
		&p.name,
		&p.description,
		&p.owner,
		&p.parent,
		&p.created,
		&p.updated,
	)

	// Tried to update a room that doesn't exist.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}

	// A ForeignKeyViolation means the referenced ownerID or parentID does not exist
	// in the rooms table, thus we will return an invalid argument error.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return nil, fmt.Errorf(
			"%s: %w: the given owner or parent given does not exist: owner '%s', parent '%s'",
			failMsg, cerrors.ErrInvalidArgument, req.Owner, req.Parent,
		)
	}

	// A UniqueViolation means the inserted room violated a uniqueness
	// constraint. The room name is not unique.
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return nil, fmt.Errorf("%s: %w: room name is not unique", failMsg, cerrors.ErrAlreadyExists)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err.Error())
	}

	return p, nil
}

func (s Service) remove(ctx context.Context, pid string) error {
	failMsg := "failed to remove room"

	log.LoggerFromContext(ctx).With("roomID", pid).Info("msg", "remove room")

	roomID, err := uuid.Parse(pid)
	if err != nil {
		return fmt.Errorf("%s: %w: invalid room id: '%s'", failMsg, cerrors.ErrInvalidArgument, pid)
	}

	_, err = s.db.ExecContext(ctx, removeQuery, roomID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", failMsg, cerrors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, cerrors.ErrInternal, err)
	}

	return nil
}
