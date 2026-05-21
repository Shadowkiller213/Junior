package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awesomeProject2/internal/model"
)

// ErrNotFound is returned when a subscription is not found.
var ErrNotFound = errors.New("subscription not found")

// ErrInvalid is returned when request data is invalid.
var ErrInvalid = errors.New("invalid request")

// Repository defines the data access methods required by the service.
type Repository interface {
	Create(ctx context.Context, s *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error)
	Update(ctx context.Context, s *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	TotalCost(ctx context.Context, startPeriod, endPeriod time.Time, userID *uuid.UUID, serviceName *string) (int64, error)
}

// Service handles subscription business logic.
type Service struct {
	repo Repository
}

// New creates a new Service with the provided repository.
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create validates and creates a new subscription.
func (s *Service) Create(ctx context.Context, req *model.CreateRequest) (*model.Subscription, error) {
	if err := validateCreate(req); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalid, err.Error())
	}

	now := time.Now().UTC()
	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("service.Create: %w", err)
	}
	return sub, nil
}

// GetByID retrieves a subscription by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.GetByID: %w", err)
	}
	if sub == nil {
		return nil, ErrNotFound
	}
	return sub, nil
}

// List retrieves subscriptions with optional filters.
func (s *Service) List(ctx context.Context, userIDStr *string, serviceNameStr *string) ([]*model.Subscription, error) {
	var userID *uuid.UUID
	if userIDStr != nil && *userIDStr != "" {
		id, err := uuid.Parse(*userIDStr)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid user_id: %s", ErrInvalid, err.Error())
		}
		userID = &id
	}

	subs, err := s.repo.List(ctx, userID, serviceNameStr)
	if err != nil {
		return nil, fmt.Errorf("service.List: %w", err)
	}
	if subs == nil {
		subs = []*model.Subscription{}
	}
	return subs, nil
}

// Update validates and updates an existing subscription.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req *model.UpdateRequest) (*model.Subscription, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	if err := validateUpdate(req); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalid, err.Error())
	}

	existing.ServiceName = req.ServiceName
	existing.Price = req.Price
	existing.UserID = req.UserID
	existing.StartDate = req.StartDate
	existing.EndDate = req.EndDate
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}
	return existing, nil
}

// Delete removes a subscription by ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("service.Delete: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service.Delete: %w", err)
	}
	return nil
}

// TotalCost calculates total subscription cost over a period with optional filters.
func (s *Service) TotalCost(
	ctx context.Context,
	startPeriodStr, endPeriodStr string,
	userIDStr *string,
	serviceNameStr *string,
) (*model.TotalResponse, error) {
	if startPeriodStr == "" {
		return nil, fmt.Errorf("%w: start_period is required", ErrInvalid)
	}
	if endPeriodStr == "" {
		return nil, fmt.Errorf("%w: end_period is required", ErrInvalid)
	}

	startMY := model.MonthYear(startPeriodStr)
	if err := startMY.Validate(); err != nil {
		return nil, fmt.Errorf("%w: invalid start_period: %s", ErrInvalid, err.Error())
	}

	endMY := model.MonthYear(endPeriodStr)
	if err := endMY.Validate(); err != nil {
		return nil, fmt.Errorf("%w: invalid end_period: %s", ErrInvalid, err.Error())
	}

	startPeriod := startMY.ToTime()
	endPeriod := endMY.ToTime()

	if endPeriod.Before(startPeriod) {
		return nil, fmt.Errorf("%w: end_period must be >= start_period", ErrInvalid)
	}

	var userID *uuid.UUID
	if userIDStr != nil && *userIDStr != "" {
		id, err := uuid.Parse(*userIDStr)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid user_id: %s", ErrInvalid, err.Error())
		}
		userID = &id
	}

	total, err := s.repo.TotalCost(ctx, startPeriod, endPeriod, userID, serviceNameStr)
	if err != nil {
		return nil, fmt.Errorf("service.TotalCost: %w", err)
	}

	return &model.TotalResponse{Total: total}, nil
}

// validateCreate validates a CreateRequest.
func validateCreate(req *model.CreateRequest) error {
	if req.ServiceName == "" {
		return errors.New("service_name is required")
	}
	if req.Price <= 0 {
		return errors.New("price must be a positive integer")
	}
	if req.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if err := req.StartDate.Validate(); err != nil {
		return fmt.Errorf("start_date: %w", err)
	}
	if req.EndDate != nil {
		if err := req.EndDate.Validate(); err != nil {
			return fmt.Errorf("end_date: %w", err)
		}
		if req.EndDate.ToTime().Before(req.StartDate.ToTime()) {
			return errors.New("end_date must be >= start_date")
		}
	}
	return nil
}

// validateUpdate validates an UpdateRequest.
func validateUpdate(req *model.UpdateRequest) error {
	if req.ServiceName == "" {
		return errors.New("service_name is required")
	}
	if req.Price <= 0 {
		return errors.New("price must be a positive integer")
	}
	if req.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if err := req.StartDate.Validate(); err != nil {
		return fmt.Errorf("start_date: %w", err)
	}
	if req.EndDate != nil {
		if err := req.EndDate.Validate(); err != nil {
			return fmt.Errorf("end_date: %w", err)
		}
		if req.EndDate.ToTime().Before(req.StartDate.ToTime()) {
			return errors.New("end_date must be >= start_date")
		}
	}
	return nil
}
