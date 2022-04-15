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

package players

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestServiceNew(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create mock db")
	}

	s := New(db)

	if s.db != db {
		t.Error("Failed to set service db")
	}
	if s.h.s != s {
		t.Error("Failed to set handler service")
	}
}

func TestServiceName(t *testing.T) {
	s := Service{}
	if s.Name() != "players" {
		t.Error("Unexpected service name")
	}
}

func TestServiceShutdown(t *testing.T) {
	s := Service{}
	s.Shutdown()
}

func TestServiceList(t *testing.T) {
}

func TestServiceGet(t *testing.T) {
}

func TestServiceCreate(t *testing.T) {
}

func TestServiceUpdate(t *testing.T) {
}

func TestServiceRemove(t *testing.T) {
}
