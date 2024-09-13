package gudu

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

// router default router configuration
func (g *Gudu) defaultRouter() http.Handler {
	mux := chi.NewRouter()
	// package built-in middlewares
	// A good base middleware stack
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	if g.DebugMode {
		mux.Use(middleware.Logger)
	}
	mux.Use(middleware.Recoverer)

	// developer default middleware
	mux.Use(g.SessionLoadAndSave)

	return mux
}
