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

package networking // import "arcadium.dev/networking"

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"arcadium.dev/core/errors"

	"arcadium.dev/arcade"
)

// NewRoomsFilter creates a RoomsFilter from the the given request's URL
// query parameters
func NewRoomsFilter(r *http.Request) (arcade.RoomsFilter, error) {
	q := r.URL.Query()
	filter := arcade.RoomsFilter{
		Limit: arcade.DefaultRoomsFilterLimit,
	}

	if values := q["ownerID"]; len(values) > 0 {
		ownerID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.RoomsFilter{}, fmt.Errorf("%w: invalid ownerID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.OwnerID = arcade.PlayerID(uuid.NullUUID{
			UUID:  ownerID,
			Valid: true,
		})
	}

	if values := q["parentID"]; len(values) > 0 {
		parentID, err := uuid.Parse(values[0])
		if err != nil {
			return arcade.RoomsFilter{}, fmt.Errorf("%w: invalid parentID query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.ParentID = arcade.RoomID(uuid.NullUUID{
			UUID:  parentID,
			Valid: true,
		})
	}

	if values := q["limit"]; len(values) > 0 {
		limit, err := strconv.Atoi(values[0])
		if err != nil || limit <= 0 || limit > arcade.MaxRoomsFilterLimit {
			return arcade.RoomsFilter{}, fmt.Errorf("%w: invalid limit query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Limit = uint(limit)
	}

	if values := q["offset"]; len(values) > 0 {
		offset, err := strconv.Atoi(values[0])
		if err != nil || offset <= 0 {
			return arcade.RoomsFilter{}, fmt.Errorf("%w: invalid offset query parameter: '%s'", errors.ErrBadRequest, values[0])
		}
		filter.Offset = uint(offset)
	}

	return filter, nil
}
