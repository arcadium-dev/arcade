package data

import (
	"context"
	"database/sql"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/core/errors"
)

type (
	// RoomStorage ...
	RoomStorage struct {
		DB     *sql.DB
		Driver RoomDriver
	}

	// RoomDriver represents the SQL driver specific functionality.
	RoomDriver interface {
		Driver
		ListQuery(assets.RoomFilter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		RemoveQuery() string
	}
)

func (i RoomStorage) List(context.Context, assets.RoomFilter) ([]*assets.Room, error) {
	return nil, errors.ErrNotImplemented
}

func (i RoomStorage) Get(context.Context, assets.RoomID) (*assets.Room, error) {
	return nil, errors.ErrNotImplemented
}

func (i RoomStorage) Create(context.Context, assets.RoomCreate) (*assets.Room, error) {
	return nil, errors.ErrNotImplemented
}

func (i RoomStorage) Update(context.Context, assets.RoomID, assets.RoomUpdate) (*assets.Room, error) {
	return nil, errors.ErrNotImplemented
}

func (i RoomStorage) Remove(context.Context, assets.RoomID) error {
	return errors.ErrNotImplemented
}
