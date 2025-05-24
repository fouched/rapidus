package rapidus

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (r *Rapidus) routes() http.Handler {
	mux := chi.NewRouter()
	addMiddleware(mux, r)

	return mux
}

func addMiddleware(mux *chi.Mux, r *Rapidus) {
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	//if r.Debug {
	//	mux.Use(middleware.Logger)
	//}

	mux.Use(r.SessionLoad)
	mux.Use(r.NoSurf)
}
