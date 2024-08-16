// Copyright 2024 arcadium.dev <info@arcadium.dev>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client // import "arcadium.dev/arcade/user/client"

import (
	"crypto/tls"
	"net/http"
	"time"
)

type (
	// Option provides options for configuring the creation of a users client.
	Option interface {
		apply(*UsersClient)
	}
)

// WithTimeout sets the client timeout.
func WithTimeout(timeout time.Duration) Option {
	return newOption(func(c *UsersClient) {
		c.httpClient.Timeout = timeout
	})
}

// WithTLSConfig sets the http client's tls config.
func WithTLSConfig(cfg *tls.Config) Option {
	return newOption(func(c *UsersClient) {
		c.httpClient.Transport = &http.Transport{
			TLSClientConfig: cfg,
		}
	})
}

type (
	option struct {
		f func(*UsersClient)
	}
)

func newOption(f func(*UsersClient)) option {
	return option{f: f}
}

func (o option) apply(c *UsersClient) {
	o.f(c)
}
