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

package data // import "arcadium.dev/arcade/asset/data"

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"arcadium.dev/core/errors"
	"arcadium.dev/core/sql"

	"arcadium.dev/arcade/user"
)

type (
	// UserStorage is used to manage the persistent storage of user data.
	UserStorage struct {
		DB     *sql.DB
		Driver UserDriver
	}

	// UserDriver abstracts away the SQL driver specific functionality.
	UserDriver interface {
		Driver
		ListQuery(user.Filter) string
		GetQuery() string
		CreateQuery() string
		UpdateQuery() string
		AssociatePlayerQuery() string
		RemoveQuery() string
	}
)

// List returns a slice of users based on the balue of the filter.
func (u UserStorage) List(ctx context.Context, filter user.Filter) ([]*user.User, error) {
	failMsg := "failed to list users"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msg("list users")

	rows, err := u.DB.Query(ctx, u.Driver.ListQuery(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Err(err).Msg("failed to close rows of list query")
		}
	}()

	users := make([]*user.User, 0)
	for rows.Next() {
		var user user.User
		err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.PublicKey,
			&user.PlayerID,
			&user.Created,
			&user.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return users, nil
}

// Get returns a user given the userID.
func (u UserStorage) Get(ctx context.Context, userID user.ID) (*user.User, error) {
	failMsg := "failed to get user"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("get user, id: '%s'", userID)

	var user user.User
	err := u.DB.QueryRow(ctx, u.Driver.GetQuery(), userID).Scan(
		&user.ID,
		&user.Login,
		&user.PublicKey,
		&user.PlayerID,
		&user.Created,
		&user.Updated,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)
	case err != nil:
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return &user, nil
}

// Create persists a new user, returning user information including the newly created userID.
func (u UserStorage) Create(ctx context.Context, create user.Create) (*user.User, error) {
	failMsg := "failed to create user"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("create user: %s", create.Login)

	var user user.User
	err := u.DB.QueryRow(ctx, u.Driver.CreateQuery(),
		create.Login,
		create.PublicKey,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PublicKey,
		&user.PlayerID,
		&user.Created,
		&user.Updated,
	)

	switch {
	// A UniqueViolation means the inserted user violated a uniqueness
	// constraint. The login is not unique.
	case u.Driver.IsUniqueViolation(err):
		return nil, fmt.Errorf("%s: %w: user login '%s' already exists", failMsg, errors.ErrConflict, create.Login)

	case err != nil:
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	logger.Info().Msgf("created user, login: '%s' id: '%s'", user.Login, user.ID)

	return &user, nil
}

// Update updates a user, return the updated user.
func (u UserStorage) Update(ctx context.Context, userID user.ID, update user.Update) (*user.User, error) {
	failMsg := "failed to update user"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("update user, id: '%s'", userID)

	var (
		user user.User
	)

	err := u.DB.QueryRow(ctx, u.Driver.UpdateQuery(),
		userID,
		update.Login,
		update.PublicKey,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PublicKey,
		&user.PlayerID,
		&user.Created,
		&user.Updated,
	)

	switch {
	// Tried to update a user that doesn't exist.
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)

	case err != nil:
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err.Error())
	}

	logger.Info().Msgf("updated user, login: '%s' id: '%s'", user.Login, user.ID)

	return &user, nil
}

// AssociatePlayer associates a player with the user, returning the updated user.
func (u UserStorage) AssociatePlayer(ctx context.Context, userID user.ID, assoc user.AssociatePlayer) (*user.User, error) {
	failMsg := "failed to associate player with user"
	logger := zerolog.Ctx(ctx)

	logger.Info().Msgf("update user, id: '%s'", userID)

	var (
		user user.User
	)

	err := u.DB.QueryRow(ctx, u.Driver.AssociatePlayerQuery(),
		userID,
		assoc.PlayerID,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PublicKey,
		&user.PlayerID,
		&user.Created,
		&user.Updated,
	)

	switch {
	// Tried to update a user that doesn't exist.
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("%s: %w", failMsg, errors.ErrNotFound)

	// A ForeignKeyViolation means the referenced playerID does not exist
	// in the players table, thus we will return an invalid argument error.
	case u.Driver.IsForeignKeyViolation(err):
		return nil, fmt.Errorf(
			"%s: %w: the given playerID does not exist, playerID: '%s'",
			failMsg, errors.ErrBadRequest, assoc.PlayerID,
		)

	case err != nil:
		return nil, fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err.Error())
	}

	logger.Info().Msgf("updated user, login: '%s' id: '%s'", user.Login, user.ID)

	return &user, nil
}

// Remove deletes the user from persistent storage.
func (u UserStorage) Remove(ctx context.Context, userID user.ID) error {
	failMsg := "failed to remove user"

	zerolog.Ctx(ctx).Info().Msgf("remove user, id: '%s'", userID)

	if _, err := u.DB.Exec(ctx, u.Driver.RemoveQuery(), userID); err != nil {
		return fmt.Errorf("%s: %w: %s", failMsg, errors.ErrInternal, err)
	}

	return nil
}
