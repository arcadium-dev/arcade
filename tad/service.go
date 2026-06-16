//  Copyright 2026 arcadium.dev <info@arcadium.dev>
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

package tad // import "arcadium.dev/arcade/tad"

import (
	"context"

	"arcadium.dev/core/telnet"
)

type (
	Service struct {
	}
)

// Name returns the name of this service.
func (s Service) Name() string {
	return "telnet_adapter"
}

func (s *Service) ServeTELNET(session *telnet.Session) {
	if err := session.WriteLine("Welcome!\n"); err != nil {
		return
	}

	for {
		line, err := session.ReadLine()
		if err != nil {
			return
		}

		if len(line) == 0 {
			if err = session.WriteLine("Goodbye!\n"); err != nil {
				return
			}
			return
		}

		if err = session.WriteLine("You wrote: " + line + "\n"); err != nil {
			return
		}
	}
}

func (s *Service) Shutdown(ctx context.Context) {
	// TODO
}
