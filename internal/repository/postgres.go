package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"awesomeProject2/internal/model"
)

// Repository handles all database operations for subscriptions.
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a new Repository with the provided connection pool.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new subscription into the database.
func (r *Repository) Create(ctx context.Context, s *model.Subscription) error {
	slog.Debug("repository: creating subscription", "id", s.ID)

	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var endDate *time.Time
	if s.EndDate != nil {
		t := s.EndDate.ToTime()
		endDate = &t
	}

	_, err := r.pool.Exec(ctx, query,
		s.ID,
		s.ServiceName,
		s.Price,
		s.UserID,
		s.StartDate.ToTime(),
		endDate,
		s.CreatedAt,
		s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}

// GetByID retrieves a subscription by its UUID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	slog.Debug("repository: getting subscription by ID", "id", id)

	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	sub, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository.GetByID: %w", err)
	}
	return sub, nil
}

// List retrieves all subscriptions with optional filters.
func (r *Repository) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	slog.Debug("repository: listing subscriptions", "user_id", userID, "service_name", serviceName)

	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1::uuid)
		  AND ($2::text IS NULL OR LOWER(service_name) = LOWER($2::text))
		ORDER BY created_at DESC
	`

	var userIDVal interface{}
	if userID != nil {
		userIDVal = *userID
	}

	var serviceNameVal interface{}
	if serviceName != nil {
		serviceNameVal = *serviceName
	}

	rows, err := r.pool.Query(ctx, query, userIDVal, serviceNameVal)
	if err != nil {
		return nil, fmt.Errorf("repository.List: %w", err)
	}
	defer rows.Close()

	var subscriptions []*model.Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.List scan: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.List rows: %w", err)
	}

	return subscriptions, nil
}

// Update modifies an existing subscription in the database.
func (r *Repository) Update(ctx context.Context, s *model.Subscription) error {
	slog.Debug("repository: updating subscription", "id", s.ID)

	query := `
		UPDATE subscriptions
		SET service_name = $2,
		    price        = $3,
		    user_id      = $4,
		    start_date   = $5,
		    end_date     = $6,
		    updated_at   = $7
		WHERE id = $1
	`

	var endDate *time.Time
	if s.EndDate != nil {
		t := s.EndDate.ToTime()
		endDate = &t
	}

	ct, err := r.pool.Exec(ctx, query,
		s.ID,
		s.ServiceName,
		s.Price,
		s.UserID,
		s.StartDate.ToTime(),
		endDate,
		s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository.Update: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("repository.Update: not found")
	}
	return nil
}

// Delete removes a subscription from the database by ID.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	slog.Debug("repository: deleting subscription", "id", id)

	query := `DELETE FROM subscriptions WHERE id = $1`

	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("repository.Delete: not found")
	}
	return nil
}

// TotalCost calculates the total cost of subscriptions in a given period with optional filters.
func (r *Repository) TotalCost(
	ctx context.Context,
	startPeriod, endPeriod time.Time,
	userID *uuid.UUID,
	serviceName *string,
) (int64, error) {
	slog.Debug("repository: calculating total cost",
		"start_period", startPeriod,
		"end_period", endPeriod,
		"user_id", userID,
		"service_name", serviceName,
	)

	query := `
		SELECT COALESCE(SUM(
		    price * (
		        (EXTRACT(YEAR FROM LEAST(COALESCE(end_date, $2::date), $2::date)) -
		         EXTRACT(YEAR FROM GREATEST(start_date, $1::date))) * 12 +
		        EXTRACT(MONTH FROM LEAST(COALESCE(end_date, $2::date), $2::date)) -
		        EXTRACT(MONTH FROM GREATEST(start_date, $1::date)) + 1
		    )
		), 0)::bigint
		FROM subscriptions
		WHERE start_date <= $2
		  AND (end_date IS NULL OR end_date >= $1)
		  AND ($3::uuid IS NULL OR user_id = $3::uuid)
		  AND ($4::text IS NULL OR LOWER(service_name) = LOWER($4::text))
	`

	var userIDVal interface{}
	if userID != nil {
		userIDVal = *userID
	}

	var serviceNameVal interface{}
	if serviceName != nil {
		serviceNameVal = *serviceName
	}

	var total int64
	err := r.pool.QueryRow(ctx, query, startPeriod, endPeriod, userIDVal, serviceNameVal).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("repository.TotalCost: %w", err)
	}
	return total, nil
}

// scanSubscription scans a row into a Subscription struct.
func scanSubscription(row pgx.Row) (*model.Subscription, error) {
	var s model.Subscription
	var startDate time.Time
	var endDate *time.Time

	err := row.Scan(
		&s.ID,
		&s.ServiceName,
		&s.Price,
		&s.UserID,
		&startDate,
		&endDate,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert DATE → MonthYear "MM-YYYY"
	s.StartDate = model.MonthYear(startDate.Format("01-2006"))
	if endDate != nil {
		my := model.MonthYear(endDate.Format("01-2006"))
		s.EndDate = &my
	}

	return &s, nil
}
