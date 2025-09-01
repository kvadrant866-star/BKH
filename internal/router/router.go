package router

import (
	"banner/internal/banner/controller"

	"github.com/go-chi/chi/v5"
)

func NewRouter(bc *controller.BannerController) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/counter", func(r chi.Router) {
		r.Get("/{id}", bc.ClickCounter)
	})

	r.Route("/stats", func(r chi.Router) {
		r.Post("/{id}", bc.GetStats)
	})

	return r
}
