package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	// Create a dummy handler that returns 201 Created
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap handler with metrics middleware
	metricsHandler := Metrics(handler)

	// Create a test request
	req := httptest.NewRequest("POST", "/test", nil)
	rr := httptest.NewRecorder()

	// Execute request
	metricsHandler.ServeHTTP(rr, req)

	// Verify the response
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())

	// Testing that the middleware doesn't crash is already a good start.
	// We could also potentially check the Prometheus registry, but that's more complex.
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rr, statusCode: 0}
	
	rw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, rw.statusCode)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
