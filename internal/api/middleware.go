package api

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// LogMiddleware returns a middleware that adds general log configuration to each context request
func LogMiddleware(ctx context.Context) StrictMiddlewareFunc {
	return func(f StrictHandlerFunc, operationID string) StrictHandlerFunc {
		return func(ctxReq context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error) {
			ctx := log.CopyFromContext(ctx, ctxReq)
			if reqID := middleware.GetReqID(ctxReq); reqID != "" {
				ctx = log.With(ctx, "req-id", reqID)
			}
			return f(ctx, w, r, args)
		}
	}
}

// BasicAuthMiddleware returns a middleware that performs an http basic authorization for endpoints configured with
// basic auth in the api spec.
// In uses the BasicAuthScopes value in context to figure if and endpoint needs authorization or not, because this
// value is injected automatically by openapi when basic auth is selected
func BasicAuthMiddleware(ctx context.Context, user, pass string) StrictMiddlewareFunc {
	return func(f StrictHandlerFunc, operationID string) StrictHandlerFunc {
		return func(ctxReq context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error) {
			if ctxReq.Value(BasicAuthScopes) != nil {
				userReq, passReq, ok := r.BasicAuth()
				if !ok {
					return nil, AuthError{err: errors.New("unauthorized")}
				}
				if subtle.ConstantTimeCompare([]byte(user), []byte(userReq)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(passReq)) != 1 {
					return nil, AuthError{errors.New("unauthorized")}
				}
			}
			return f(ctx, w, r, args)
		}
	}
}
