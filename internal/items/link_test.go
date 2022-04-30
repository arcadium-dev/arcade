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

package items

import (
	"testing"
	"time"

	"arcadium.dev/arcade/internal/arcade"
)

const (
	id          = "00000000-0000-0000-0000-000000000001"
	name        = "Nothing"
	description = "An item of no importance."
	owner       = "00000000-0000-0000-0000-000000000001"
	location    = "00000000-0000-0000-0000-000000000001"
	inventory   = "00000000-0000-0000-0000-000000000001"
)

var (
	created = time.Now()
	updated = time.Now()

	p = item{
		id:          id,
		name:        name,
		description: description,
		owner:       owner,
		location:    location,
		inventory:   inventory,
		created:     created,
		updated:     updated,
	}

	req = itemRequest{
		Name:        name,
		Description: description,
		Owner:       owner,
		Location:    location,
		Inventory:   inventory,
	}
)

func TestNewItem(t *testing.T) {
	p := newItem(req)
	if p.Name() != name ||
		p.Description() != description ||
		p.Owner() != owner ||
		p.Location() != location ||
		p.Inventory() != inventory {
		t.Errorf("Unexpected item: %+v", p)
	}

	if p.ID() != "" ||
		!p.Created().IsZero() ||
		!p.Updated().IsZero() {
		t.Errorf("Unexpected item: %+v", p)
	}
}

func TestItem(t *testing.T) {
	if p.ID() != id ||
		p.Name() != name ||
		p.Description() != description ||
		p.Owner() != owner ||
		p.Location() != location ||
		p.Inventory() != inventory ||
		!created.Equal(p.Created()) ||
		!updated.Equal(p.Updated()) {
		t.Errorf("Unexpected item: %+v", p)
	}
}

func TestNewItemResponsData(t *testing.T) {
	data := newItemResponseData(p)

	if data.ItemID != id ||
		data.Name != name ||
		data.Description != description ||
		data.Owner != owner ||
		data.Location != location ||
		data.Inventory != inventory ||
		!created.Equal(data.Created) ||
		!updated.Equal(data.Updated) {
		t.Errorf("Unexpected data: %+v", data)
	}
}

func TestNewItemResponse(t *testing.T) {
	r := newItemResponse(p)

	if r.Data.ItemID != id ||
		r.Data.Name != name ||
		r.Data.Description != description ||
		r.Data.Owner != owner ||
		r.Data.Location != location ||
		r.Data.Inventory != inventory ||
		!created.Equal(r.Data.Created) ||
		!updated.Equal(r.Data.Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}

func TestNewItemsReponse(t *testing.T) {
	r := newItemsResponse([]arcade.Item{p})

	if r.Data[0].ItemID != id ||
		r.Data[0].Name != name ||
		r.Data[0].Description != description ||
		r.Data[0].Owner != owner ||
		r.Data[0].Location != location ||
		r.Data[0].Inventory != inventory ||
		!created.Equal(r.Data[0].Created) ||
		!updated.Equal(r.Data[0].Updated) {
		t.Errorf("Unexpected response: %+v", r)
	}
}
