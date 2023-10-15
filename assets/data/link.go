package data

import (
	"context"
	"database/sql"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/core/errors"
)

type (
	// LinkStorage ...
	LinkStorage struct {
		DB     *sql.DB
		Driver LinkDriver
	}

	// LinkDriver represents the SQL driver specific functionality.
	LinkDriver interface {
		Driver
		ListQuery(assets.LinkFilter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		RemoveQuery() string
	}
)

func (i LinkStorage) List(context.Context, assets.LinkFilter) ([]*assets.Link, error) {
	return nil, errors.ErrNotImplemented
}

func (i LinkStorage) Get(context.Context, assets.LinkID) (*assets.Link, error) {
	return nil, errors.ErrNotImplemented
}

func (i LinkStorage) Create(context.Context, assets.LinkCreate) (*assets.Link, error) {
	return nil, errors.ErrNotImplemented
}

func (i LinkStorage) Update(context.Context, assets.LinkID, assets.LinkUpdate) (*assets.Link, error) {
	return nil, errors.ErrNotImplemented
}

func (i LinkStorage) Remove(context.Context, assets.LinkID) error {
	return errors.ErrNotImplemented
}
