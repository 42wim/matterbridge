package gitter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
)

var (
	mux    *http.ServeMux
	gitter *Gitter
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	gitter = New("abc")

	// Fake the API and Stream base URLs by using the test
	// server URL instead.
	url, _ := url.Parse(server.URL)
	gitter.config.apiBaseURL = url.String() + "/"
	gitter.config.streamBaseURL = url.String() + "/"
}

func teardown() {
	server.Close()
}
