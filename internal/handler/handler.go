package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AVZotov/gophermart/internal/auth"
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

type orderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// centsToAccrual converts a nullable cents amount (exact integer arithmetic
// used internally) to the float64 the JSON API contract expects.
func centsToAccrual(cents *int64) *float64 {
	if cents == nil {
		return nil
	}
	accrual := float64(*cents) / 100
	return &accrual
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

// userID extracts the authenticated user's ID from r's context. On failure
// it writes a 500 (context missing the user ID indicates a middleware
// wiring bug, not a client error) and returns ok=false.
func (h *Handler) userID(w http.ResponseWriter, r *http.Request) (id int64, ok bool) {
	id, ok = auth.UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("user id missing from context", "path", r.URL.Path)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
	return id, ok
}

// writeJSON sets the JSON content type, writes status, and encodes v,
// logging (but not surfacing to the client) any encode failure.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error("failed to encode response", "err", err)
	}
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

	w.Header().Set("Authorization", "Bearer "+token)
	h.writeJSON(w, http.StatusOK, authResponse{Token: token})
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

	w.Header().Set("Authorization", "Bearer "+token)
	h.writeJSON(w, http.StatusOK, authResponse{Token: token})
}

func (h *Handler) uploadOrder(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "text/plain") {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	err = h.svc.UploadOrder(r.Context(), userID, orderNumber)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
	case errors.Is(err, domain.ErrOrderAlreadyUploaded):
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, domain.ErrOrderOwnedByAnotherUser):
		http.Error(w, "order owned by another user", http.StatusConflict)
	case errors.Is(err, domain.ErrInvalidOrderID):
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
	default:
		h.logger.Error("upload order failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}

	orders, err := h.svc.GetOrders(r.Context(), userID)
	if err != nil {
		h.logger.Error("get orders failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(
			resp, orderResponse{
				Number:     o.Number,
				Status:     string(o.Status),
				Accrual:    centsToAccrual(o.AccrualCents),
				UploadedAt: o.UploadedAt,
			},
		)
	}

	h.writeJSON(w, http.StatusOK, resp)
}
