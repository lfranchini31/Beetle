// Copyright 2020 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pkg

import (
	"net/http"
	"net/http/httptest"
)

// MockServer mocks http server
func MockServer(uri, response string) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
	})

	srv := httptest.NewServer(handler)

	return srv
}
