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
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type (
	// Service that reports the metrics of the service.
	Service struct{}
)

// Register sets up the http handler for this service with the given router.
func (Service) Register(router *mux.Router) {
	router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
}

// Name returns the name of the service.
func (Service) Name() string { return "metrics" }

// Shutdown is a no-op since there are no long running processes.
func (Service) Shutdown() {}
