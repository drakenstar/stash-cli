package stash

import (
	"net/http"
	"net/http/httptest"
)

type mockEndpoint string

func (m mockEndpoint) Do(*http.Request) (*http.Response, error) {
	rw := httptest.NewRecorder()
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte(m))

	return rw.Result(), nil
}
