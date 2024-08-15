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

package rest // import "arcadium.dev/arcade/user/rest"

import "arcadium.dev/arcade"

type (
	// UserCreateRequest is used to request a user be created.
	UserCreateRequest struct {
		UserRequest
	}

	// UserUpdateRequest is used to request a user be updated.
	UserUpdateRequest struct {
		UserRequest
	}

	// UserRequest is used to request a user be created or updated.
	UserRequest struct {
		// Login is the login id of the user.
		Login string `json:"login"`

		// PublicKey is the ssh public key of the user.
		PublicKey string `json:"publicKey"`

		// PlayerID is the ID of the player associated with the user.
		PlayerID string `json:"playerID"`
	}

	// UserResponse returns a user.
	UserResponse struct {
		// User returns the information about a user.
		User User `json:"user"`
	}

	// UsersResponse returns multiple users.
	UsersResponse struct {
		// Users returns the information about multiple users.
		Users []User `json:"users"`
	}

	// User holds a user's information, and is sent in a response.
	User struct {
		// ID is the user identifier.
		ID string `json:"id"`

		// Name is the user name.
		Login string `json:"login"`

		// PublicKey is the user's ssh public key.
		PublicKey string `json:"publicKey"`

		// PlayerID is the PlayerID associated with the user.
		PlayerID string `json:"playerID"`

		// Created is the time of the user's creation.
		Created arcade.Timestamp `json:"created"`

		// Updated is the time the user was last updated.
		Updated arcade.Timestamp `json:"updated"`
	}
)
