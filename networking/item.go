//  Copyright 2022-2023 arcadium.dev <info@arcadium.dev>
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

package networking // import "arcadium.dev/networking"

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"arcadium.dev/core/errors"

	"arcadium.dev/arcade"
)

// NewItemsFilter creates an ItemFilter from the the given request's URL
// query parameters.
func NewItemsFilter(r *http.Request) (arcade.ItemsFilter, error) {
	q := r.URL.Query()
	filter := arcade.ItemsFilter{
		Limit: arcade.DefaultItemsFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = arcade.PlayerID(uuid.NullUUID{
			UUID:  ownerID,
			Valid: true,
		})
	}

	var (
		err          error
		locationUUID *uuid.UUID
	)
	if values := q["locationID"]; len(values) > 0 {
		*locationUUID, err = uuid.Parse(values[0])
		if err != nil {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
	}

	if locationUUID != nil {
		var location arcade.ItemLocationID
		if values := q["locationType"]; len(values) > 0 {
			switch strings.ToLower(values[0]) {
			case "room":
				location = arcade.ItemID(uuid.NullUUID{
					UUID:  *locationUUID,
					Valid: true,
				})
			case "player":
				location = arcade.PlayerID(uuid.NullUUID{
					UUID:  *locationUUID,
					Valid: true,
				})
			case "item":
				location = arcade.ItemID(uuid.NullUUID{
					UUID:  *locationUUID,
					Valid: true,
				})
			default:
				return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid locationType query parameter: '%s'", errors.ErrBadRequest, values[0])
			}
		}
		filter.LocationID = location
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > arcade.MaxItemsFilterLimit {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return arcade.ItemsFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	return filter, nil
}
