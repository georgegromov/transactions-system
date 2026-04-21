package http

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type RouterDeps struct {
	TransactionsHandler *transactionsHandler
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

		// Transactions routes
		r.Route("/transactions", func(r chi.Router) {
			r.Post("/", deps.TransactionsHandler.CreateTransaction)
		})
	})

	return r
}
