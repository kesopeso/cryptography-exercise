package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/status", func(r chi.Router) {
		r.Get("/", listStatuses)

		r.Post("/", createStatus)

		r.Route("/{statusId}", func(r chi.Router) {
			r.Post("/", createState)

			r.Route("/{index}", func(r chi.Router) {
				r.Get("/", getState)

				r.Put("/", setState)

				r.Delete("/", deleteState)
			})
		})
	})

	return r
}
