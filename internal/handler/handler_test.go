package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AVZotov/gophermart/internal/auth"
	"github.com/AVZotov/gophermart/internal/domain"
	"github.com/AVZotov/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopLogger struct{}

func (noopLogger) Debug(msg string, args ...interface{}) {}
func (noopLogger) Info(msg string, args ...interface{})  {}
func (noopLogger) Warn(msg string, args ...interface{})  {}
func (noopLogger) Error(msg string, args ...interface{}) {}
func (noopLogger) Sync() error                           { return nil }

type fakeStorage struct {
	createOrderFunc       func(ctx context.Context, order *domain.Order, userID int64) error
	getOrdersByUserIDFunc func(ctx context.Context, userID int64) ([]*domain.Order, error)
}

func (f *fakeStorage) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	return 0, nil
}

func (f *fakeStorage) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeStorage) CreateOrder(ctx context.Context, order *domain.Order, userID int64) error {
	return f.createOrderFunc(ctx, order, userID)
}

func (f *fakeStorage) GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return f.getOrdersByUserIDFunc(ctx, userID)
}

func (f *fakeStorage) Close() {}

func newTestHandler(store *fakeStorage) *Handler {
	svc := service.NewService(store, []byte("test-secret"))
	return NewHandler(svc, noopLogger{})
}

func withUserID(req *http.Request, userID int64) *http.Request {
	return req.WithContext(auth.WithUserID(req.Context(), userID))
}

func TestHandler_UploadOrder_Success(t *testing.T) {
	store := &fakeStorage{
		createOrderFunc: func(ctx context.Context, order *domain.Order, userID int64) error {
			return nil
		},
	}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("79927398713"))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.uploadOrder(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
}

func TestHandler_UploadOrder_AlreadyUploadedByOwner(t *testing.T) {
	store := &fakeStorage{
		createOrderFunc: func(ctx context.Context, order *domain.Order, userID int64) error {
			return domain.ErrOrderAlreadyUploaded
		},
	}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("79927398713"))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.uploadOrder(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_UploadOrder_OwnedByAnotherUser(t *testing.T) {
	store := &fakeStorage{
		createOrderFunc: func(ctx context.Context, order *domain.Order, userID int64) error {
			return domain.ErrOrderOwnedByAnotherUser
		},
	}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("79927398713"))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.uploadOrder(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandler_UploadOrder_InvalidNumber(t *testing.T) {
	store := &fakeStorage{
		createOrderFunc: func(ctx context.Context, order *domain.Order, userID int64) error {
			t.Fatal("storage should not be called for an invalid order number")
			return nil
		},
	}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678900"))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.uploadOrder(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestHandler_UploadOrder_EmptyBody(t *testing.T) {
	store := &fakeStorage{}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(""))
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.uploadOrder(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_UploadOrder_MissingUserID(t *testing.T) {
	store := &fakeStorage{}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("79927398713"))
	rec := httptest.NewRecorder()

	h.uploadOrder(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_GetOrders_Success(t *testing.T) {
	store := &fakeStorage{
		getOrdersByUserIDFunc: func(ctx context.Context, userID int64) ([]*domain.Order, error) {
			return []*domain.Order{{Number: "79927398713", Status: domain.OrderStatusNew}}, nil
		},
	}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.getOrders(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "79927398713")
}

func TestHandler_GetOrders_NoContent(t *testing.T) {
	store := &fakeStorage{
		getOrdersByUserIDFunc: func(ctx context.Context, userID int64) ([]*domain.Order, error) {
			return nil, nil
		},
	}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.getOrders(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_GetOrders_MissingUserID(t *testing.T) {
	store := &fakeStorage{}
	h := newTestHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rec := httptest.NewRecorder()

	h.getOrders(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
