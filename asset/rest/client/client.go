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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"arcadium.dev/core/errors"
	"github.com/rs/zerolog"
)

const (
	defaultTimeout = 10 * time.Second
)

type (
	// Client provides a client for the asset api.
	Client struct {
		baseURL   string
		timeout   time.Duration
		transport *http.Transport
		client    *http.Client
	}

	ResponseError struct {
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}
)

func (e ResponseError) Error() string { return e.Detail }

// New returns a new client for the asset API.
func New(baseURL string, opts ...ClientOption) *Client {
	// Set defaults.
	c := &Client{
		baseURL: baseURL,
		timeout: defaultTimeout,
	}

	// Load options.
	for _, opt := range opts {
		opt.apply(c)
	}

	c.client = &http.Client{
		Timeout: c.timeout,
	}
	if c.transport != nil {
		c.client.Transport = c.transport
	}

	return c
}

func (c Client) Send(ctx context.Context, req *http.Request) (*http.Response, error) {
	zerolog.Ctx(ctx).Debug().Msgf("sending request: %s", req.URL)

	// TODO:  Add auth to request?

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to read request body: %w", err)
		}
		defer resp.Body.Close()

		var e ResponseError
		if err := json.Unmarshal(body, &e); err == nil {
			var (
				httpErr errors.HTTPError
				ok      bool
			)
			if httpErr, ok = errors.HTTPErrors[e.Status]; !ok {
				httpErr = errors.ErrInternal
			}
			return nil, fmt.Errorf("%w, (server error: %w)", httpErr, e)
		}
		return nil, fmt.Errorf("%d, %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return resp, nil
}
