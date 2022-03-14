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
)

const (
	playerID    = "ba168738-0492-4d0e-a70d-e79ed433055b"
	name        = "Drunen"
	description = "Enjoying life"
	home        = "0a51d4e8-93c3-426b-9434-ac0849980b52"
	location    = "b57e4cbe-39c6-4254-81b6-c4e237a7ab6c"
)

var (
	created = time.Now()
	updated = time.Now()

	p = player{
		playerID:    playerID,
		name:        name,
		description: description,
		home:        home,
		location:    location,
		created:     created,
		updated:     updated,
	}

	req = playerRequest{
		PlayerID:    playerID,
		Name:        name,
		Description: description,
		Home:        home,
		Location:    location,
	}
)

func TestNewPlayerResponseData(t *testing.T) {
	data := newPlayerResponseData(p)

	if data.PlayerID != playerID ||
		data.Name != name ||
		data.Description != description ||
		data.Home != home ||
		data.Location != location ||
		data.Created != created ||
		data.Updated != updated {
		t.Errorf("Unexpected data: %+v", data)
	}
}

func TestNewPlayerReponse(t *testing.T) {
	r := newPlayerResponse(p)

	if r.Data.PlayerID != playerID ||
		r.Data.Name != name ||
		r.Data.Description != description ||
		r.Data.Home != home ||
		r.Data.Location != location ||
		r.Data.Created != created ||
		r.Data.Updated != updated {
		t.Errorf("Unexpected response: %+v", r)
	}
}

func TestNewPlayersResponse(t *testing.T) {
	r := newPlayersResponse([]player{p})

	if len(r.Data) != 1 {
		t.Errorf("Unexpected response: %+v", r)
	}
	if r.Data[0].PlayerID != playerID ||
		r.Data[0].Name != name ||
		r.Data[0].Description != description ||
		r.Data[0].Home != home ||
		r.Data[0].Location != location ||
		r.Data[0].Created != created ||
		r.Data[0].Updated != updated {
		t.Errorf("Unexpected response: %+v", r)
	}
}

func TestNewPlayer(t *testing.T) {
	p := newPlayer(req)

	if p.playerID != playerID ||
		p.name != name ||
		p.description != description ||
		p.home != home ||
		p.location != location {
		t.Errorf("Unexpected player: %+v", p)
	}
}

func TestPlayer(t *testing.T) {
	if p.PlayerID() != playerID ||
		p.Name() != name ||
		p.Description() != description ||
		p.Home() != home ||
		p.Location() != location ||
		p.Created() != created ||
		p.Updated() != updated {
		t.Errorf("Unexpected player: %+v", p)
	}
}
