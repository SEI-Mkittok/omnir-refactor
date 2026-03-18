package handler_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/middleware"
)

// withClaims injects JWT claims into the request context, mimicking the Authenticate middleware.
func withClaims(r *http.Request, claims *auth.Claims) *http.Request {
	return r.WithContext(middleware.WithClaims(r.Context(), claims))
}

// withURLParam injects a Chi URL param into a request context.
func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// nopLogger returns a logger that discards all output, suitable for tests.
func nopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
