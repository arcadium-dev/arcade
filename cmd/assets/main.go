// Copyright 2021-2023 arcadium.dev <info@arcadium.dev>
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

package main

import (
	l "log"

	"arcadium.dev/core/http/server"
	"arcadium.dev/core/rest"
)

var (
	Version string
	Branch  string
	Commit  string
	Date    string
)

type (
	// RestServer defines the expected behavior of a rest server.
	RestServer interface {
		Init() error
		Start(...server.Service) error
	}
)

// New creates a new rest server. This is provided as a function variable to
// allow for easier unit testing.
var New = func(v, b, c, d string) RestServer {
	return rest.NewServer(v, b, c, d)
}

// Main is the testable entry point into the assets server.
func Main() error {
	server := New(Version, Branch, Commit, Date)
	if err := server.Init(); err != nil {
		return err
	}

	// TODO: Setup assets service here.

	if err := server.Start(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		l.Fatal(err)
	}
}
