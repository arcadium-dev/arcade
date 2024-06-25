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

package postgres_test

/*
import (
	"errors"
	"fmt"
	"testing"

	"arcadium.dev/arcade"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"arcadium.dev/arcade/storage/cockroach"
)

func TestDriver(t *testing.T) {
	d := cockroach.Driver{}

	if d.PlayersGetQuery() != cockroach.PlayersGetQuery {
		t.Error("query mismatch")
	}
	if d.PlayersCreateQuery() != cockroach.PlayersCreateQuery {
		t.Error("query mismatch")
	}
	if d.PlayersUpdateQuery() != cockroach.PlayersUpdateQuery {
		t.Error("query mismatch")
	}
	if d.PlayersRemoveQuery() != cockroach.PlayersRemoveQuery {
		t.Error("query mismatch")
	}

	if d.RoomsListQuery(arcade.RoomsFilter{}) != cockroach.RoomsListQuery {
		t.Error("query mismatch")
	}
	if d.RoomsGetQuery() != cockroach.RoomsGetQuery {
		t.Error("query mismatch")
	}
	if d.RoomsCreateQuery() != cockroach.RoomsCreateQuery {
		t.Error("query mismatch")
	}
	if d.RoomsUpdateQuery() != cockroach.RoomsUpdateQuery {
		t.Error("query mismatch")
	}
	if d.RoomsRemoveQuery() != cockroach.RoomsRemoveQuery {
		t.Error("query mismatch")
	}

	if d.LinksListQuery(arcade.LinksFilter{}) != cockroach.LinksListQuery {
		t.Error("query mismatch")
	}
	if d.LinksGetQuery() != cockroach.LinksGetQuery {
		t.Error("query mismatch")
	}
	if d.LinksCreateQuery() != cockroach.LinksCreateQuery {
		t.Error("query mismatch")
	}
	if d.LinksUpdateQuery() != cockroach.LinksUpdateQuery {
		t.Error("query mismatch")
	}
	if d.LinksRemoveQuery() != cockroach.LinksRemoveQuery {
		t.Error("query mismatch")
	}

	if d.ItemsListQuery(arcade.ItemsFilter{}) != cockroach.ItemsListQuery {
		t.Error("query mismatch")
	}
	if d.ItemsGetQuery() != cockroach.ItemsGetQuery {
		t.Error("query mismatch")
	}
	if d.ItemsCreateQuery() != cockroach.ItemsCreateQuery {
		t.Error("query mismatch")
	}
	if d.ItemsUpdateQuery() != cockroach.ItemsUpdateQuery {
		t.Error("query mismatch")
	}
	if d.ItemsRemoveQuery() != cockroach.ItemsRemoveQuery {
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

func TestPlayersListQuery(t *testing.T) {
	d := cockroach.Driver{}

	filter := arcade.PlayersFilter{}

	actual := d.PlayersListQuery(filter)
	expected := cockroach.PlayersListQuery
	if expected != actual {
		t.Errorf("\nExpected query: %s\nActual query   %s", expected, actual)
	}

	id := uuid.New()
	filter.LocationID = &id
	actual = d.PlayersListQuery(filter)
	expected = cockroach.PlayersListQuery + fmt.Sprintf(" WHERE location_id = '%s'", id)
	if expected != actual {
		t.Errorf("\nExpected query: %s\nActual query:   %s", expected, actual)
	}

	limit := 42
	filter.LocationID = nil
	filter.Limit = limit
	actual = d.PlayersListQuery(filter)
	expected = cockroach.PlayersListQuery + fmt.Sprintf(" LIMIT %d", limit)
	if expected != actual {
		t.Errorf("\nExpected query: %s\nActual query:   %s", expected, actual)
	}

	offset := 10
	filter.Limit = 0
	filter.Offset = offset
	actual = d.PlayersListQuery(filter)
	expected = cockroach.PlayersListQuery + fmt.Sprintf(" OFFSET %d", offset)
	if expected != actual {
		t.Errorf("\nExpected query: %s\nActual query:   %s", expected, actual)
	}

	filter.LocationID = &id
	filter.Limit = limit
	filter.Offset = offset
	actual = d.PlayersListQuery(filter)
	expected = cockroach.PlayersListQuery + fmt.Sprintf(" WHERE location_id = '%s' LIMIT %d OFFSET %d", id, limit, offset)
	if expected != actual {
		t.Errorf("\nExpected query: %s\nActual query:   %s", expected, actual)
	}
}
*/
