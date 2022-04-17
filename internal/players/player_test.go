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

package players

import (
	"testing"
	"time"

	"arcadium.dev/arcade/internal/arcade"
)

const (
	id          = "00000000-0000-0000-0000-000000000001"
	name        = "Nobody"
	description = "A person of no importance."
	home        = "00000000-0000-0000-0000-000000000001"
	location    = "00000000-0000-0000-0000-000000000001"
)

var (
	created = time.Now()
	updated = time.Now()

	p = player{
		id:          id,
		name:        name,
		description: description,
		home:        home,
		location:    location,
		created:     created,
		updated:     updated,
	}

	req = playerRequest{
		Name:        name,
		Description: description,
		Home:        home,
		Location:    location,
	}
)

func TestNewPlayer(t *testing.T) {
	p := newPlayer(req)
	if p.Name() != name ||
		p.Description() != description ||
		p.Home() != home ||
		p.Location() != location {
		t.Errorf("Unexpected player: %+v", p)
	}

	if p.ID() != "" ||
		!p.Created().IsZero() ||
		!p.Updated().IsZero() {
		t.Errorf("Unexpected player: %+v", p)
	}
}

func TestPlayer(t *testing.T) {
	if p.ID() != id ||
		p.Name() != name ||
		p.Description() != description ||
		p.Home() != home ||
		p.Location() != location ||
		!created.Equal(p.Created()) ||
		!updated.Equal(p.Updated()) {
		t.Errorf("Unexpected player: %+v", p)
	}
}

func TestNewPlayerResponsData(t *testing.T) {
	data := newPlayerResponseData(p)

	if data.PlayerID != id ||
		data.Name != name ||
		data.Description != description ||
		data.Home != home ||
		data.Location != location ||
		!created.Equal(data.Created) ||
		!updated.Equal(data.Updated) {
		t.Errorf("Unexpected data: %+v", data)
	}
}

func TestNewPlayerResponse(t *testing.T) {
	r := newPlayerResponse(p)

	if r.Data.PlayerID != id ||
		r.Data.Name != name ||
		r.Data.Description != description ||
		r.Data.Home != home ||
		r.Data.Location != location ||
		!created.Equal(r.Data.Created) ||
		!updated.Equal(r.Data.Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}

func TestNewPlayersReponse(t *testing.T) {
	r := newPlayersResponse([]arcade.Player{p})

	if r.Data[0].PlayerID != id ||
		r.Data[0].Name != name ||
		r.Data[0].Description != description ||
		r.Data[0].Home != home ||
		r.Data[0].Location != location ||
		!created.Equal(r.Data[0].Created) ||
		!updated.Equal(r.Data[0].Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}
