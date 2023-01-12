package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// LogMiddleware returns a middleware that adds general log configuration to each context request
func LogMiddleware(ctx context.Context) StrictMiddlewareFunc {
	return func(f StrictHandlerFunc, operationID string) StrictHandlerFunc {
		return func(ctxReq context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error) {
			return f(log.With(log.CopyFromContext(ctx, ctxReq), "req-id", middleware.GetReqID(ctxReq)), w, r, args)
		}
	}
}
