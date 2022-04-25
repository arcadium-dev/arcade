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
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or impliep.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package rooms

import (
	"testing"
	"time"

	"arcadium.dev/arcade/internal/arcade"
)

const (
	id          = "00000000-0000-0000-0000-000000000001"
	name        = "Nobody"
	description = "A person of no importance."
	owner       = "00000000-0000-0000-0000-000000000001"
	parent      = "00000000-0000-0000-0000-000000000001"
)

var (
	created = time.Now()
	updated = time.Now()

	p = room{
		id:          id,
		name:        name,
		description: description,
		owner:       owner,
		parent:      parent,
		created:     created,
		updated:     updated,
	}

	req = roomRequest{
		Name:        name,
		Description: description,
		Owner:       owner,
		Parent:      parent,
	}
)

func TestNewRoom(t *testing.T) {
	p := newRoom(req)
	if p.Name() != name ||
		p.Description() != description ||
		p.Owner() != owner ||
		p.Parent() != parent {
		t.Errorf("Unexpected room: %+v", p)
	}

	if p.ID() != "" ||
		!p.Created().IsZero() ||
		!p.Updated().IsZero() {
		t.Errorf("Unexpected room: %+v", p)
	}
}

func TestRoom(t *testing.T) {
	if p.ID() != id ||
		p.Name() != name ||
		p.Description() != description ||
		p.Owner() != owner ||
		p.Parent() != parent ||
		!created.Equal(p.Created()) ||
		!updated.Equal(p.Updated()) {
		t.Errorf("Unexpected room: %+v", p)
	}
}

func TestNewRoomResponsData(t *testing.T) {
	data := newRoomResponseData(p)

	if data.RoomID != id ||
		data.Name != name ||
		data.Description != description ||
		data.Owner != owner ||
		data.Parent != parent ||
		!created.Equal(data.Created) ||
		!updated.Equal(data.Updated) {
		t.Errorf("Unexpected data: %+v", data)
	}
}

func TestNewRoomResponse(t *testing.T) {
	r := newRoomResponse(p)

	if r.Data.RoomID != id ||
		r.Data.Name != name ||
		r.Data.Description != description ||
		r.Data.Owner != owner ||
		r.Data.Parent != parent ||
		!created.Equal(r.Data.Created) ||
		!updated.Equal(r.Data.Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}

func TestNewRoomsReponse(t *testing.T) {
	r := newRoomsResponse([]arcade.Room{p})

	if r.Data[0].RoomID != id ||
		r.Data[0].Name != name ||
		r.Data[0].Description != description ||
		r.Data[0].Owner != owner ||
		r.Data[0].Parent != parent ||
		!created.Equal(r.Data[0].Created) ||
		!updated.Equal(r.Data[0].Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}
