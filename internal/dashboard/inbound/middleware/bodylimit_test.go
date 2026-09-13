package middleware_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bodyProbe struct {
	handler http.Handler
	read    int64
	readErr error
}

func newBodyProbe() *bodyProbe {
	probe := &bodyProbe{}
	probe.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := io.Copy(io.Discard, r.Body)
		probe.read = n
		probe.readErr = err
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return probe
}

func oversizedRequest(method, path string) *http.Request {
	const twoMiB = 2 << 20
	return httptest.NewRequest(method, path, bytes.NewReader(make([]byte, twoMiB)))
}

func TestBodyLimit(t *testing.T) {
	t.Run("caps an oversized body at the default 1 MiB", func(t *testing.T) {
		probe := newBodyProbe()
		handler := middleware.BodyLimit(middleware.DefaultBodyLimitBytes)(probe.handler)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, oversizedRequest(http.MethodPost, "/api/actions"))

		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
		var maxErr *http.MaxBytesError
		require.ErrorAs(t, probe.readErr, &maxErr)
		assert.Equal(t, int64(middleware.DefaultBodyLimitBytes), maxErr.Limit)
		assert.LessOrEqual(t, probe.read, int64(middleware.DefaultBodyLimitBytes)+1)
	})

	t.Run("allows a 2 MiB restore body under the 32 MiB rule", func(t *testing.T) {
		probe := newBodyProbe()
		handler := middleware.BodyLimit(middleware.DefaultBodyLimitBytes, middleware.DefaultBodyLimitRules()...)(probe.handler)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, oversizedRequest(http.MethodPost, "/api/restore"))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NoError(t, probe.readErr)
		assert.Equal(t, int64(2<<20), probe.read)
	})

	t.Run("allows a 2 MiB sh3d import body under the 64 MiB rule", func(t *testing.T) {
		probe := newBodyProbe()
		handler := middleware.BodyLimit(middleware.DefaultBodyLimitBytes, middleware.DefaultBodyLimitRules()...)(probe.handler)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, oversizedRequest(http.MethodPost, "/api/levels/import/sh3d"))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NoError(t, probe.readErr)
		assert.Equal(t, int64(2<<20), probe.read)
	})

	t.Run("allows a 2 MiB plan import body under the suffix rule", func(t *testing.T) {
		probe := newBodyProbe()
		handler := middleware.BodyLimit(middleware.DefaultBodyLimitBytes, middleware.DefaultBodyLimitRules()...)(probe.handler)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, oversizedRequest(http.MethodPost, "/api/levels/42/plan/import"))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NoError(t, probe.readErr)
		assert.Equal(t, int64(2<<20), probe.read)
	})

	t.Run("keeps the default limit for unmatched paths", func(t *testing.T) {
		probe := newBodyProbe()
		handler := middleware.BodyLimit(middleware.DefaultBodyLimitBytes, middleware.DefaultBodyLimitRules()...)(probe.handler)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, oversizedRequest(http.MethodPost, "/api/devices"))

		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
		var maxErr *http.MaxBytesError
		require.ErrorAs(t, probe.readErr, &maxErr)
		assert.Equal(t, int64(middleware.DefaultBodyLimitBytes), maxErr.Limit)
	})
}

func TestDefaultBodyLimit(t *testing.T) {
	probe := newBodyProbe()
	handler := middleware.DefaultBodyLimit()(probe.handler)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, oversizedRequest(http.MethodPost, "/api/restore"))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, probe.readErr)
	assert.Equal(t, int64(2<<20), probe.read)
}
