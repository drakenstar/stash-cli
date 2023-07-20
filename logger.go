package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type loggingTransport struct{}

func (s *loggingTransport) Do(r *http.Request) (*http.Response, error) {
	bytes, _ := httputil.DumpRequestOut(r, true)
	resp, err := http.DefaultTransport.RoundTrip(r)
	respBytes, _ := httputil.DumpResponse(resp, true)
	bytes = append(bytes, respBytes...)
	fmt.Printf("%s\n", bytes)
	return resp, err
}
