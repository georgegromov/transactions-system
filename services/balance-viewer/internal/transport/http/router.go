package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type RouterDeps struct {
	BalanceHandler *balanceHandler
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {

		// Balance routes
		r.Route("/account", func(r chi.Router) {
			r.Get("/{account_id}/balance", deps.BalanceHandler.GetAccountBalance)
		})
	})

	return r
}
