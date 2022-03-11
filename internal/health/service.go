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

package health

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	route string = "/health"
)

type (
	// Service reports on the health of the service as a whole.
	Service struct{}
)

// Register sets up the http handler for this service with the given router.
func (s Service) Register(router *mux.Router) {
	r := router.PathPrefix(route).Subrouter()
	r.HandleFunc("", s.get).Methods(http.MethodGet)
}

// Name returns the name of the service.
func (Service) Name() string {
	return "health"
}

// Shutdown is a no-op since there no long running processes for this service.
func (Service) Shutdown() {}

type (
	response struct {
		Data responseData `json:"data"`
	}
	responseData struct {
		Status string `json:"status"`
	}
)

func (Service) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Data: responseData{Status: "up"}})
}
