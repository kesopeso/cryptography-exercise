package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kesopeso/cryptography-exercise/internal/store"
)

// NewRouter creates a chi router with logging, middleware,
// and all status API routes registered.
func NewRouter(statusStore store.StatusStore, keyPath string, authToken string) *chi.Mux {
	r := chi.NewRouter()
	h := newStatusHandlers(statusStore, keyPath)
	authMiddleware := bearerAuth(authToken)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/status", func(r chi.Router) {
		r.Get("/", h.getStatusIds)

		r.With(authMiddleware).Post("/", h.createStatus)

		r.Route("/{statusId}", func(r chi.Router) {
			r.With(authMiddleware).Post("/", h.createStatusValue)

			r.Route("/{index}", func(r chi.Router) {
				r.Get("/", h.getStatusValue)

				r.With(authMiddleware).Put("/", h.updateStatusValueToTrue)

				r.With(authMiddleware).Delete("/", h.updateStatusValueToFalse)
			})
		})
	})

	return r
}
