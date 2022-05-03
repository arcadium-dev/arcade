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

package arcade // import "arcadium.dev/arcade"

type (
	// Storage represents the SQL driver specific functionality.
	StorageDriver interface {
		// PlayersListQuery returns the List query string given the filter.
		PlayersListQuery(PlayersFilter) string

		// PlayersGetQuery returns the Get query string.
		PlayersGetQuery() string

		// PlayersCreateQuery returns the Create query string.
		PlayersCreateQuery() string

		// PlayersUpdateQuery returns the update query string.
		PlayersUpdateQuery() string

		// PlayersRemoveQuery returns the Remove query string.
		PlayersRemoveQuery() string

		// RoomListQuery returns the List query string given the filter.
		RoomsListQuery(RoomsFilter) string

		// RoomsGetQuery returns the Get query string.
		RoomsGetQuery() string

		// RoomsCreateQuery returns the Create query string.
		RoomsCreateQuery() string

		// RoomsUpdateQuery returns the Update query string.
		RoomsUpdateQuery() string

		// RoomsRemoveQuery returns the Remove query string.
		RoomsRemoveQuery() string

		// LinksListQuery returns the List query string given the filter.
		LinksListQuery(LinksFilter) string

		// LinksGetQuery returns the Get query string.
		LinksGetQuery() string

		// LinksCreateQuery returns the Create query string.
		LinksCreateQuery() string

		// LinksUpdateQuery returns the Update query string.
		LinksUpdateQuery() string

		// LinksRemoveQuery returns the Remove query string.
		LinksRemoveQuery() string

		// ItemsListQuery returns the List query string given the filter.
		ItemsListQuery(ItemsFilter) string

		// ItemsGetQuery returns the Get query string.
		ItemsGetQuery() string

		// ItemsCreateQuery returns the Create query string.
		ItemsCreateQuery() string

		// ItemsUpdateQuery returns the Update query string.
		ItemsUpdateQuery() string

		// ItemsRemoveQuery returns the Remove query string.
		ItemsRemoveQuery() string

		// IsForeignKeyViolation returns true if the given error is a foreign key violation error.
		IsForeignKeyViolation(err error) bool

		// IsUniqueViolation returns true if the given error is a unique violation error.
		IsUniqueViolation(err error) bool
	}
)
