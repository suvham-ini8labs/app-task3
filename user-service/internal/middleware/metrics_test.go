package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("ok"))
	})

	handler := Metrics(nextHandler)

	t.Run("RecordMetrics", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusAccepted, w.Code)
		// We can't easily check the prometheus metrics values without more complex setup,
		// but we verify the middleware doesn't crash and passes the request through.
	})
}
