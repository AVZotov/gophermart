package handler

import (
	"github.com/AVZotov/gophermart/internal/auth"
	"github.com/AVZotov/gophermart/internal/logger"
	"github.com/AVZotov/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// NewRouter builds the chi.Mux for the API, wiring request logging and the
// public and authenticated (JWT-protected, via secret) route groups.
func NewRouter(h *Handler, l logger.Logger, secret []byte) *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.LoggingMiddleware(l))
	register(mux, h, secret)
	return mux
}

func register(mux *chi.Mux, h *Handler, secret []byte) {
	mux.Post("/api/user/register", h.register)
	mux.Post("/api/user/login", h.login)

	mux.Group(
		func(mux chi.Router) {
			mux.Use(auth.AuthMiddleware(secret))
		},
	)
}
