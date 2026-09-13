package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// TestHandlerCoversRouterRedirects is the regression test for the audit finding "Security headers
// are absent on router path-normalisation redirects": gorilla/mux answers those 301s inside its own
// ServeHTTP, before any router.Use middleware runs, so the headers have to be applied by the
// outermost handler the server is given.
func TestHandlerCoversRouterRedirects(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/b", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	server := httptest.NewServer(newHandler(router))
	defer server.Close()

	// "//b" is cleaned to "/b" by the router, which answers with a redirect of its own. The
	// redirect must not be followed: the headers on the 301 itself are what is under test.
	client := *server.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Get(server.URL + "//b")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
	require.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	require.NotEmpty(t, resp.Header.Get("Content-Security-Policy"))
	require.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
}
