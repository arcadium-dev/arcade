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

	"github.com/google/uuid"

	"arcadium.dev/core/errors"

	"arcadium.dev/arcade"
)

// NewPlayersFilter creates a PlayersFilter from the the given request's URL
// query parameters.
func NewPlayersFilter(r *http.Request) (arcade.PlayersFilter, error) {
	q := r.URL.Query()
	filter := arcade.PlayersFilter{
		Limit: arcade.DefaultPlayersFilterLimit,
	}

	if values := q["locationID"]; len(values) > 0 {
		locationID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.PlayersFilter{}, fmt.Errorf("%w: invalid locationID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.LocationID = arcade.RoomID(uuid.NullUUID{
			UUID:  locationID,
			Valid: true,
		})
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > arcade.MaxPlayersFilterLimit {
			return arcade.PlayersFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return arcade.PlayersFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	return filter, nil
}
