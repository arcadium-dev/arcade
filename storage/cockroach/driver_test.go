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

package cockroach

import (
	"errors"
	"testing"

	"arcadium.dev/arcade"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

func TestDriver(t *testing.T) {
	d := Driver{}

	if d.PlayersListQuery(arcade.PlayersFilter{}) != playersListQuery {
		t.Error("query mismatch")
	}
	if d.PlayersGetQuery() != playersGetQuery {
		t.Error("query mismatch")
	}
	if d.PlayersCreateQuery() != playersCreateQuery {
		t.Error("query mismatch")
	}
	if d.PlayersUpdateQuery() != playersUpdateQuery {
		t.Error("query mismatch")
	}
	if d.PlayersRemoveQuery() != playersRemoveQuery {
		t.Error("query mismatch")
	}

	if d.RoomsListQuery(arcade.RoomsFilter{}) != roomsListQuery {
		t.Error("query mismatch")
	}
	if d.RoomsGetQuery() != roomsGetQuery {
		t.Error("query mismatch")
	}
	if d.RoomsCreateQuery() != roomsCreateQuery {
		t.Error("query mismatch")
	}
	if d.RoomsUpdateQuery() != roomsUpdateQuery {
		t.Error("query mismatch")
	}
	if d.RoomsRemoveQuery() != roomsRemoveQuery {
		t.Error("query mismatch")
	}

	if d.LinksListQuery(arcade.LinksFilter{}) != linksListQuery {
		t.Error("query mismatch")
	}
	if d.LinksGetQuery() != linksGetQuery {
		t.Error("query mismatch")
	}
	if d.LinksCreateQuery() != linksCreateQuery {
		t.Error("query mismatch")
	}
	if d.LinksUpdateQuery() != linksUpdateQuery {
		t.Error("query mismatch")
	}
	if d.LinksRemoveQuery() != linksRemoveQuery {
		t.Error("query mismatch")
	}

	if d.ItemsListQuery(arcade.ItemsFilter{}) != itemsListQuery {
		t.Error("query mismatch")
	}
	if d.ItemsGetQuery() != itemsGetQuery {
		t.Error("query mismatch")
	}
	if d.ItemsCreateQuery() != itemsCreateQuery {
		t.Error("query mismatch")
	}
	if d.ItemsUpdateQuery() != itemsUpdateQuery {
		t.Error("query mismatch")
	}
	if d.ItemsRemoveQuery() != itemsRemoveQuery {
		t.Error("query mismatch")
	}

	if d.IsForeignKeyViolation(errors.New("nope")) {
		t.Error("huh?")
	}
	err := &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}
	if !d.IsForeignKeyViolation(err) {
		t.Error("foreign key error expected")
	}

	if d.IsUniqueViolation(errors.New("nope")) {
		t.Error("huh?")
	}
	err = &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	if !d.IsUniqueViolation(err) {
		t.Error("unique error expected")
	}
}
