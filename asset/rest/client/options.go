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

package client // import "arcadium.dev/arcade/asset/rest/client"

import (
	"crypto/tls"
	"net/http"
	"time"
)

type (
	// ClientOption provides options for configuring the creation of an asset client.
	ClientOption interface {
		apply(*Client)
	}
)

type (
	clientOption struct {
		f func(*Client)
	}
)

// WithTime sets the client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return newClientOption(func(c *Client) {
		if timeout > 0 {
			c.timeout = timeout
		}
	})
}

// WithInsecure ... temporary
func WithInsecure() ClientOption {
	return newClientOption(func(c *Client) {
		c.transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	})
}

func newClientOption(f func(*Client)) clientOption {
	return clientOption{f: f}
}

func (o clientOption) apply(c *Client) {
	o.f(c)
}
