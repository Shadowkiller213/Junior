package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"awesomeProject2/internal/model"
	"awesomeProject2/internal/service"
)

// mockRepo is a test double for the Repository interface.
type mockRepo struct {
	createFn    func(ctx context.Context, s *model.Subscription) error
	getByIDFn   func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	listFn      func(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error)
	updateFn    func(ctx context.Context, s *model.Subscription) error
	deleteFn    func(ctx context.Context, id uuid.UUID) error
	totalCostFn func(ctx context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) (int64, error)
}

func (m *mockRepo) Create(ctx context.Context, s *model.Subscription) error {
	if m.createFn != nil {
		return m.createFn(ctx, s)
	}
	return nil
}
func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockRepo) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, serviceName)
	}
	return nil, nil
}
func (m *mockRepo) Update(ctx context.Context, s *model.Subscription) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, s)
	}
	return nil
}
func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockRepo) TotalCost(ctx context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) (int64, error) {
	if m.totalCostFn != nil {
		return m.totalCostFn(ctx, start, end, userID, serviceName)
	}
	return 0, nil
}

// helpers

func fixedSub() *model.Subscription {
	return &model.Subscription{
		ID:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   "07-2025",
	}
}

func fixedCreateReq() *model.CreateRequest {
	return &model.CreateRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   "07-2025",
	}
}

// --- Create ---

func TestService_Create_OK(t *testing.T) {
	svc := service.New(&mockRepo{})
	sub, err := svc.Create(context.Background(), fixedCreateReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.ServiceName != "Yandex Plus" {
		t.Errorf("ServiceName = %q, want %q", sub.ServiceName, "Yandex Plus")
	}
	if sub.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
}

func TestService_Create_MissingServiceName(t *testing.T) {
	svc := service.New(&mockRepo{})
	req := fixedCreateReq()
	req.ServiceName = ""
	_, err := svc.Create(context.Background(), req)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_Create_NegativePrice(t *testing.T) {
	svc := service.New(&mockRepo{})
	req := fixedCreateReq()
	req.Price = -1
	_, err := svc.Create(context.Background(), req)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_Create_ZeroPrice(t *testing.T) {
	svc := service.New(&mockRepo{})
	req := fixedCreateReq()
	req.Price = 0
	_, err := svc.Create(context.Background(), req)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_Create_InvalidStartDate(t *testing.T) {
	svc := service.New(&mockRepo{})
	req := fixedCreateReq()
	req.StartDate = "13-2025"
	_, err := svc.Create(context.Background(), req)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_Create_EndDateBeforeStartDate(t *testing.T) {
	svc := service.New(&mockRepo{})
	req := fixedCreateReq()
	end := model.MonthYear("06-2025")
	req.EndDate = &end
	_, err := svc.Create(context.Background(), req)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_Create_RepoError(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *model.Subscription) error {
			return errors.New("db down")
		},
	}
	svc := service.New(repo)
	_, err := svc.Create(context.Background(), fixedCreateReq())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- GetByID ---

func TestService_GetByID_Found(t *testing.T) {
	sub := fixedSub()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) {
			return sub, nil
		},
	}
	svc := service.New(repo)
	got, err := svc.GetByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != sub.ID {
		t.Errorf("ID = %v, want %v", got.ID, sub.ID)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) {
			return nil, nil
		},
	}
	svc := service.New(repo)
	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, service.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- List ---

func TestService_List_NoFilters(t *testing.T) {
	subs := []*model.Subscription{fixedSub()}
	repo := &mockRepo{
		listFn: func(_ context.Context, _ *uuid.UUID, _ *string) ([]*model.Subscription, error) {
			return subs, nil
		},
	}
	svc := service.New(repo)
	got, err := svc.List(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestService_List_InvalidUserID(t *testing.T) {
	svc := service.New(&mockRepo{})
	bad := "not-a-uuid"
	_, err := svc.List(context.Background(), &bad, nil)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_List_Empty(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, _ *uuid.UUID, _ *string) ([]*model.Subscription, error) {
			return nil, nil
		},
	}
	svc := service.New(repo)
	got, err := svc.List(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Error("expected empty slice, got nil")
	}
}

// --- Update ---

func TestService_Update_OK(t *testing.T) {
	sub := fixedSub()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) { return sub, nil },
		updateFn:  func(_ context.Context, _ *model.Subscription) error { return nil },
	}
	svc := service.New(repo)
	req := &model.UpdateRequest{
		ServiceName: "Netflix",
		Price:       799,
		UserID:      sub.UserID,
		StartDate:   "01-2025",
	}
	got, err := svc.Update(context.Background(), sub.ID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ServiceName != "Netflix" {
		t.Errorf("ServiceName = %q, want %q", got.ServiceName, "Netflix")
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) { return nil, nil },
	}
	svc := service.New(repo)
	req := &model.UpdateRequest{
		ServiceName: "Netflix",
		Price:       799,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}
	_, err := svc.Update(context.Background(), uuid.New(), req)
	if !errors.Is(err, service.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- Delete ---

func TestService_Delete_OK(t *testing.T) {
	sub := fixedSub()
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) { return sub, nil },
		deleteFn:  func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	svc := service.New(repo)
	if err := svc.Delete(context.Background(), sub.ID); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Subscription, error) { return nil, nil },
	}
	svc := service.New(repo)
	err := svc.Delete(context.Background(), uuid.New())
	if !errors.Is(err, service.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- TotalCost ---

func TestService_TotalCost_OK(t *testing.T) {
	repo := &mockRepo{
		totalCostFn: func(_ context.Context, _, _ time.Time, _ *uuid.UUID, _ *string) (int64, error) {
			return 1200, nil
		},
	}
	svc := service.New(repo)
	res, err := svc.TotalCost(context.Background(), "01-2025", "03-2025", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 1200 {
		t.Errorf("Total = %d, want 1200", res.Total)
	}
}

func TestService_TotalCost_MissingStart(t *testing.T) {
	svc := service.New(&mockRepo{})
	_, err := svc.TotalCost(context.Background(), "", "03-2025", nil, nil)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_TotalCost_MissingEnd(t *testing.T) {
	svc := service.New(&mockRepo{})
	_, err := svc.TotalCost(context.Background(), "01-2025", "", nil, nil)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_TotalCost_EndBeforeStart(t *testing.T) {
	svc := service.New(&mockRepo{})
	_, err := svc.TotalCost(context.Background(), "06-2025", "01-2025", nil, nil)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_TotalCost_InvalidPeriod(t *testing.T) {
	svc := service.New(&mockRepo{})
	_, err := svc.TotalCost(context.Background(), "13-2025", "03-2025", nil, nil)
	if !errors.Is(err, service.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}
