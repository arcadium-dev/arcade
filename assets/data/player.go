package data

import (
	"context"
	"database/sql"

	"arcadium.dev/arcade/assets"
	"arcadium.dev/core/errors"
)

type (
	// PlayerStorage ...
	PlayerStorage struct {
		DB     *sql.DB
		Driver PlayerDriver
	}

	// PlayerDriver represents the SQL driver specific functionality.
	PlayerDriver interface {
		Driver
		ListQuery(assets.PlayersFilter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		RemoveQuery() string
	}
)

func (i PlayerStorage) List(context.Context, assets.PlayersFilter) ([]*assets.Player, error) {
	return nil, errors.ErrNotImplemented
}

func (i PlayerStorage) Get(context.Context, assets.PlayerID) (*assets.Player, error) {
	return nil, errors.ErrNotImplemented
}

func (i PlayerStorage) Create(context.Context, assets.PlayerCreate) (*assets.Player, error) {
	return nil, errors.ErrNotImplemented
}

func (i PlayerStorage) Update(context.Context, assets.PlayerID, assets.PlayerUpdate) (*assets.Player, error) {
	return nil, errors.ErrNotImplemented
}

func (i PlayerStorage) Remove(context.Context, assets.PlayerID) error {
	return errors.ErrNotImplemented
}
