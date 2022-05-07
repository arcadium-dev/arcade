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

package http_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"arcadium.dev/core/http"
)

func checkRespError(t *testing.T, w *httptest.ResponseRecorder, status int, errMsg string) {
	t.Helper()

	resp := w.Result()
	if resp.StatusCode != status {
		t.Errorf("Unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response body")
	}
	defer resp.Body.Close()

	var respErr struct {
		Error http.ResponseError `json:"error,omitempty"`
	}
	err = json.Unmarshal(body, &respErr)
	if err != nil {
		t.Errorf("Failed to json unmarshal error response: %s", err)
	}

	if !strings.Contains(respErr.Error.Detail, errMsg) {
		t.Errorf("\nExpected error detail: %s\nActual error detail:   %s", errMsg, respErr.Error.Detail)
	}
	if respErr.Error.Status != status {
		t.Errorf("Unexpected error status: %d", respErr.Error.Status)
	}
}
