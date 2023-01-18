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
			ctx := log.CopyFromContext(ctx, ctxReq)
			if reqID := middleware.GetReqID(ctxReq); reqID != "" {
				ctx = log.With(ctx, "req-id", reqID)
			}
			return f(ctx, w, r, args)
		}
	}
}
