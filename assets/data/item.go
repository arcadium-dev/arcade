package data

import (
	"context"
	"database/sql"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/core/errors"
)

type (
	// ItemStorage ...
	ItemStorage struct {
		DB     *sql.DB
		Driver ItemDriver
	}

	// ItemDriver represents the SQL driver specific functionality.
	ItemDriver interface {
		Driver
		ListQuery(assets.ItemsFilter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		RemoveQuery() string
	}

	Driver interface {
		IsForeignKeyViolation(err error) bool
		IsUniqueViolation(err error) bool
	}
)

func (i ItemStorage) List(context.Context, assets.ItemsFilter) ([]*assets.Item, error) {
	return nil, errors.ErrNotImplemented
}

func (i ItemStorage) Get(context.Context, assets.ItemID) (*assets.Item, error) {
	return nil, errors.ErrNotImplemented
}

func (i ItemStorage) Create(context.Context, assets.ItemCreate) (*assets.Item, error) {
	return nil, errors.ErrNotImplemented
}

func (i ItemStorage) Update(context.Context, assets.ItemID, assets.ItemUpdate) (*assets.Item, error) {
	return nil, errors.ErrNotImplemented
}

func (i ItemStorage) Remove(context.Context, assets.ItemID) error {
	return errors.ErrNotImplemented
}
