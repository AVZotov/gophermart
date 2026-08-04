package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AVZotov/gophermart/internal/domain"
	"github.com/AVZotov/gophermart/internal/logger"
	"github.com/AVZotov/gophermart/internal/service"
)

type registerRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

// maxRequestBodyBytes bounds the size of decoded request bodies to guard
// against unbounded memory allocation from oversized payloads.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// Handler holds the HTTP handlers for the API, delegating all business
// logic to the wrapped service.
type Handler struct {
	svc    *service.Service
	logger logger.Logger
}

// NewHandler constructs a Handler backed by svc for business logic and l
// for logging.
func NewHandler(svc *service.Service, l logger.Logger) *Handler {
	return &Handler{svc: svc, logger: l}
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.svc.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserExists):
			http.Error(w, "login already taken", http.StatusConflict)
		case errors.Is(err, domain.ErrPasswordTooLong):
			http.Error(w, "password too long", http.StatusBadRequest)
		default:
			h.logger.Error("register failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(authResponse{Token: token}); err != nil {
		h.logger.Error("failed to encode response", "err", err)
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.svc.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			http.Error(w, "invalid login or password", http.StatusUnauthorized)
		default:
			h.logger.Error("login failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(authResponse{Token: token}); err != nil {
		h.logger.Error("failed to encode response", "err", err)
	}
}
