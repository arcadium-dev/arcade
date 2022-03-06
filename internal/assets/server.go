//  Copyright 2021-2022 arcadium.dev <info@arcadium.dev>
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

package assets

// Build information.
var (
	Name    string
	Version string
	Branch  string
	Commit  string
	Date    string
	Go      string
)

type (
	// Server represents the arcade server.
	Server struct{}
)

// New returns a new arcade server.
func New(name, version, branch, commit, date, gover string) *Server {
	Name = name
	Version = version
	Branch = branch
	Commit = commit
	Date = date
	Go = gover

	s := &Server{}

	return s
}

// Start is the entry point into the service.
func (s *Server) Start(args []string) {
}

// Stop halts the server.
func (s *Server) Stop() {
}
