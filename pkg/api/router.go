package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetRouter initialises a new http router and applies all routes
func GetRouter() http.Handler {
	r := chi.NewRouter()
	return applyRoutes(r)
}

func applyRoutes(r chi.Router) chi.Router {
	r.Route("/", func(r chi.Router) {
		r.Get("/", getIndex)
	})

	return r
}
