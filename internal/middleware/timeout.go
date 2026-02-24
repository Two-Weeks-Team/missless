package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout wraps a handler with a context deadline.
// If the handler exceeds the given duration, the context is cancelled.
func Timeout(d time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
