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

package items

import (
	"time"

	"arcadium.dev/arcade/internal/arcade"
)

const (
	maxNameLen        = 255
	maxDescriptionLen = 4096
)

type (
	// item is the internal representation of the data related to an item.
	item struct {
		id          string
		name        string
		description string
		owner       string
		location    string
		inventory   string
		created     time.Time
		updated     time.Time
	}

	// itemRequest is the payload of an item request.
	itemRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       string `json:"owner"`
		Location    string `json:"location"`
		Inventory   string `json:"inventory"`
	}

	// itemResponse is used as payload data for item responses.
	itemResponseData struct {
		ItemID      string    `json:"itemID"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Owner       string    `json:"owner"`
		Location    string    `json:"location"`
		Inventory   string    `json:"inventory"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
	}

	// itemResponse is used to json encoded a response with a single item.
	itemResponse struct {
		Data itemResponseData `json:"data"`
	}

	// itemsResponse is used to json encoded a response with a multiple items.
	itemsResponse struct {
		Data []itemResponseData `json:"data"`
	}
)

func newItem(p itemRequest) arcade.Item {
	return item{
		name:        p.Name,
		description: p.Description,
		owner:       p.Owner,
		location:    p.Location,
		inventory:   p.Inventory,
	}
}

func (p item) ID() string          { return p.id }
func (p item) Name() string        { return p.name }
func (p item) Description() string { return p.description }
func (p item) Owner() string       { return p.owner }
func (p item) Location() string    { return p.location }
func (p item) Inventory() string   { return p.inventory }
func (p item) Created() time.Time  { return p.created }
func (p item) Updated() time.Time  { return p.updated }

func newItemResponseData(p arcade.Item) itemResponseData {
	return itemResponseData{
		ItemID:      p.ID(),
		Name:        p.Name(),
		Description: p.Description(),
		Owner:       p.Owner(),
		Location:    p.Location(),
		Inventory:   p.Inventory(),
		Created:     p.Created(),
		Updated:     p.Updated(),
	}
}

func newItemResponse(p arcade.Item) itemResponse {
	return itemResponse{
		Data: newItemResponseData(p),
	}
}

func newItemsResponse(ps []arcade.Item) itemsResponse {
	var r itemsResponse
	for _, p := range ps {
		r.Data = append(r.Data, newItemResponseData(p))
	}
	return r
}
