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

package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestServiceRegister(t *testing.T) {
	method := http.MethodGet
	route := "/metrics"

	router := mux.NewRouter()
	s := Service{}
	s.Register(router)

	r := httptest.NewRequest(method, route, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body")
	}
	defer resp.Body.Close()

	if len(body) == 0 {
		t.Error("Expected a response body")
	}
}

func TestServiceName(t *testing.T) {
	var s Service
	if s.Name() != "metrics" {
		t.Errorf("Unexpected service name: %s", s.Name())
	}
}
