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

package arcade // import "arcadium.dev/arcade"

import (
	"errors"
	"fmt"
	"time"
)

const (
	TimestampFormat = "2006-01-02T15:04:05.999999"
)

type (
	Timestamp struct {
		time.Time
	}
)

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", t.Time.UTC().Format(TimestampFormat))), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	l := len(b)
	if l <= 2 || b[0] != byte('"') || b[l-1] != byte('"') {
		return errors.New("failed to unmarshal timestamp, invalid timestamp")
	}

	var err error
	if ts, err := time.Parse(TimestampFormat, string(b[1:l-1])); err == nil {
		*t = Timestamp{Time: ts}
	}
	return err
}
