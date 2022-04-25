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

package links

import (
	"time"

	"arcadium.dev/arcade/internal/arcade"
)

const (
	maxNameLen        = 255
	maxDescriptionLen = 4096
)

type (
	// link is the internal representation of the data related to a link.
	link struct {
		id          string
		name        string
		description string
		owner       string
		location    string
		destination string
		created     time.Time
		updated     time.Time
	}

	// linkRequest is the payload of a link request.
	linkRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       string `json:"owner"`
		Location    string `json:"location"`
		Destination string `json:"destination"`
	}

	// linkResponse is used as payload data for link responses.
	linkResponseData struct {
		LinkID      string    `json:"linkID"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Owner       string    `json:"owner"`
		Location    string    `json:"location"`
		Destination string    `json:"destination"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
	}

	// linkResponse is used to json encoded a response with a single link.
	linkResponse struct {
		Data linkResponseData `json:"data"`
	}

	// linksResponse is used to json encoded a response with a multiple links.
	linksResponse struct {
		Data []linkResponseData `json:"data"`
	}
)

func newLink(p linkRequest) arcade.Link {
	return link{
		name:        p.Name,
		description: p.Description,
		owner:       p.Owner,
		location:    p.Location,
		destination: p.Destination,
	}
}

func (p link) ID() string          { return p.id }
func (p link) Name() string        { return p.name }
func (p link) Description() string { return p.description }
func (p link) Owner() string       { return p.owner }
func (p link) Location() string    { return p.location }
func (p link) Destination() string { return p.destination }
func (p link) Created() time.Time  { return p.created }
func (p link) Updated() time.Time  { return p.updated }

func newLinkResponseData(p arcade.Link) linkResponseData {
	return linkResponseData{
		LinkID:      p.ID(),
		Name:        p.Name(),
		Description: p.Description(),
		Owner:       p.Owner(),
		Location:    p.Location(),
		Destination: p.Destination(),
		Created:     p.Created(),
		Updated:     p.Updated(),
	}
}

func newLinkResponse(p arcade.Link) linkResponse {
	return linkResponse{
		Data: newLinkResponseData(p),
	}
}

func newLinksResponse(ps []arcade.Link) linksResponse {
	var r linksResponse
	for _, p := range ps {
		r.Data = append(r.Data, newLinkResponseData(p))
	}
	return r
}
