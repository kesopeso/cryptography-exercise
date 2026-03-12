package server

import (
	"net/http"
	"strings"
)

// bearerAuth returns a middleware that requires a valid Bearer token in the
// Authorization header.
func bearerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			bearer, found := strings.CutPrefix(header, "Bearer ")
			if !found || bearer != token {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
