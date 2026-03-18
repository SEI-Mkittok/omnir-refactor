package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const slowRequestThreshold = 500 * time.Millisecond

// Logger returns a Chi-compatible structured logging middleware.
// It logs duration_ms as a numeric field for aggregation and warns on slow requests.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				elapsed := time.Since(start)
				attrs := []any{
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration_ms", float64(elapsed.Microseconds()) / 1000.0,
					"request_id", middleware.GetReqID(r.Context()),
					"remote_addr", r.RemoteAddr,
				}
				if elapsed >= slowRequestThreshold {
					logger.Warn("slow request", attrs...)
				} else {
					logger.Info("request", attrs...)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
