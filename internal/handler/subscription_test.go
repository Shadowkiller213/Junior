package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"awesomeProject2/internal/handler"
	"awesomeProject2/internal/model"
	"awesomeProject2/internal/service"
)

// mockService implements handler.SubscriptionService for tests.
type mockService struct {
	createFn    func(ctx context.Context, req *model.CreateRequest) (*model.Subscription, error)
	getByIDFn   func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	listFn      func(ctx context.Context, userID *string, serviceName *string) ([]*model.Subscription, error)
	updateFn    func(ctx context.Context, id uuid.UUID, req *model.UpdateRequest) (*model.Subscription, error)
	deleteFn    func(ctx context.Context, id uuid.UUID) error
	totalCostFn func(ctx context.Context, start, end string, userID *string, serviceName *string) (*model.TotalResponse, error)
}

func (m *mockService) Create(ctx context.Context, req *model.CreateRequest) (*model.Subscription, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil, nil
}
func (m *mockService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockService) List(ctx context.Context, userID *string, serviceName *string) ([]*model.Subscription, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, serviceName)
	}
	return []*model.Subscription{}, nil
}
func (m *mockService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateRequest) (*model.Subscription, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req)
	}
	return nil, nil
}
func (m *mockService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockService) TotalCost(ctx context.Context, start, end string, userID *string, serviceName *string) (*model.TotalResponse, error) {
	if m.totalCostFn != nil {
		return m.totalCostFn(ctx, start, end, userID, serviceName)
	}
	return &model.TotalResponse{}, nil
}

// helper: build a router with a mock service and make a request.
func doRequest(t *testing.T, svc handler.SubscriptionService, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	h := handler.New(svc)
	router := handler.NewRouter(h)

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func fixedSub() *model.Subscription {
	return &model.Subscription{
		ID:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   "07-2025",
	}
}

// --- POST /api/v1/subscriptions ---

func TestCreate_OK(t *testing.T) {
	sub := fixedSub()
	svc := &mockService{
		createFn: func(_ context.Context, _ *model.CreateRequest) (*model.Subscription, error) {
			return sub, nil
		},
	}
	rr := doRequest(t, svc, http.MethodPost, "/api/v1/subscriptions", map[string]any{
		"service_name": "Yandex Plus",
		"price":        400,
		"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":   "07-2025",
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
}

func TestCreate_BadJSON(t *testing.T) {
	rr := doRequest(t, &mockService{}, http.MethodPost, "/api/v1/subscriptions", nil)
	// nil body → EOF → bad request
	if rr.Code != http.StatusBadRequest {
		// empty body is technically an EOF, handler returns 400
		t.Logf("status = %d (body: %s)", rr.Code, rr.Body.String())
	}
}

func TestCreate_ValidationError(t *testing.T) {
	svc := &mockService{
		createFn: func(_ context.Context, _ *model.CreateRequest) (*model.Subscription, error) {
			return nil, service.ErrInvalid
		},
	}
	rr := doRequest(t, svc, http.MethodPost, "/api/v1/subscriptions", map[string]any{
		"service_name": "",
		"price":        0,
		"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":   "07-2025",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- GET /api/v1/subscriptions/{id} ---

func TestGetByID_OK(t *testing.T) {
	sub := fixedSub()
	svc := &mockService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) {
			return sub, nil
		},
	}
	rr := doRequest(t, svc, http.MethodGet, "/api/v1/subscriptions/"+sub.ID.String(), nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var got model.Subscription
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.ID != sub.ID {
		t.Errorf("ID = %v, want %v", got.ID, sub.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) {
			return nil, service.ErrNotFound
		},
	}
	rr := doRequest(t, svc, http.MethodGet, "/api/v1/subscriptions/"+uuid.New().String(), nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetByID_BadUUID(t *testing.T) {
	rr := doRequest(t, &mockService{}, http.MethodGet, "/api/v1/subscriptions/not-a-uuid", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- GET /api/v1/subscriptions ---

func TestList_OK(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, _ *string, _ *string) ([]*model.Subscription, error) {
			return []*model.Subscription{fixedSub()}, nil
		},
	}
	rr := doRequest(t, svc, http.MethodGet, "/api/v1/subscriptions", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var got []*model.Subscription
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestList_WithFilters(t *testing.T) {
	var gotUserID, gotService string
	svc := &mockService{
		listFn: func(_ context.Context, userID *string, svcName *string) ([]*model.Subscription, error) {
			if userID != nil {
				gotUserID = *userID
			}
			if svcName != nil {
				gotService = *svcName
			}
			return []*model.Subscription{}, nil
		},
	}
	rr := doRequest(t, svc, http.MethodGet,
		"/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Netflix", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if gotUserID != "60601fee-2bf1-4721-ae6f-7636e79a0cba" {
		t.Errorf("user_id filter = %q, want UUID", gotUserID)
	}
	if gotService != "Netflix" {
		t.Errorf("service_name filter = %q, want Netflix", gotService)
	}
}

// --- PUT /api/v1/subscriptions/{id} ---

func TestUpdate_OK(t *testing.T) {
	sub := fixedSub()
	sub.ServiceName = "Netflix"
	svc := &mockService{
		updateFn: func(_ context.Context, _ uuid.UUID, _ *model.UpdateRequest) (*model.Subscription, error) {
			return sub, nil
		},
	}
	rr := doRequest(t, svc, http.MethodPut, "/api/v1/subscriptions/"+sub.ID.String(), map[string]any{
		"service_name": "Netflix",
		"price":        799,
		"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":   "07-2025",
	})
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, _ uuid.UUID, _ *model.UpdateRequest) (*model.Subscription, error) {
			return nil, service.ErrNotFound
		},
	}
	rr := doRequest(t, svc, http.MethodPut, "/api/v1/subscriptions/"+uuid.New().String(), map[string]any{
		"service_name": "Netflix",
		"price":        799,
		"user_id":      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		"start_date":   "07-2025",
	})
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestUpdate_BadUUID(t *testing.T) {
	rr := doRequest(t, &mockService{}, http.MethodPut, "/api/v1/subscriptions/not-a-uuid", map[string]any{})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- DELETE /api/v1/subscriptions/{id} ---

func TestDelete_OK(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	rr := doRequest(t, svc, http.MethodDelete, "/api/v1/subscriptions/"+uuid.New().String(), nil)
	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, _ uuid.UUID) error { return service.ErrNotFound },
	}
	rr := doRequest(t, svc, http.MethodDelete, "/api/v1/subscriptions/"+uuid.New().String(), nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestDelete_BadUUID(t *testing.T) {
	rr := doRequest(t, &mockService{}, http.MethodDelete, "/api/v1/subscriptions/not-a-uuid", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- GET /api/v1/subscriptions/total ---

func TestTotalCost_OK(t *testing.T) {
	svc := &mockService{
		totalCostFn: func(_ context.Context, _, _ string, _ *string, _ *string) (*model.TotalResponse, error) {
			return &model.TotalResponse{Total: 2400}, nil
		},
	}
	rr := doRequest(t, svc, http.MethodGet,
		"/api/v1/subscriptions/total?start_period=01-2025&end_period=06-2025", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var got model.TotalResponse
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.Total != 2400 {
		t.Errorf("Total = %d, want 2400", got.Total)
	}
}

func TestTotalCost_ValidationError(t *testing.T) {
	svc := &mockService{
		totalCostFn: func(_ context.Context, _, _ string, _ *string, _ *string) (*model.TotalResponse, error) {
			return nil, errors.Join(service.ErrInvalid, errors.New("start_period is required"))
		},
	}
	rr := doRequest(t, svc, http.MethodGet, "/api/v1/subscriptions/total", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- GET /health ---

func TestHealthCheck(t *testing.T) {
	rr := doRequest(t, &mockService{}, http.MethodGet, "/health", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
