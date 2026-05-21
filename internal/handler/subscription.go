package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"awesomeProject2/internal/model"
	"awesomeProject2/internal/service"
)

// SubscriptionService defines what the handler needs from the service layer.
type SubscriptionService interface {
	Create(ctx context.Context, req *model.CreateRequest) (*model.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, userID *string, serviceName *string) ([]*model.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateRequest) (*model.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	TotalCost(ctx context.Context, startPeriod, endPeriod string, userID *string, serviceName *string) (*model.TotalResponse, error)
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	svc SubscriptionService
}

// New creates a new Handler.
func New(svc SubscriptionService) *Handler {
	return &Handler{svc: svc}
}

// errorResponse writes a JSON error payload.
func errorResponse(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// writeJSON writes a JSON response with a given status code.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// Create godoc
//
//	@Summary		Create a subscription
//	@Description	Create a new subscription record
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.CreateRequest	true	"Subscription data"
//	@Success		201		{object}	model.Subscription
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/subscriptions [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("handler.Create: bad request body", "error", err)
		errorResponse(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	sub, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalid) {
			slog.Warn("handler.Create: validation error", "error", err)
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("handler.Create: internal error", "error", err)
		errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("handler.Create: subscription created", "id", sub.ID)
	writeJSON(w, http.StatusCreated, sub)
}

// GetByID godoc
//
//	@Summary		Get a subscription by ID
//	@Description	Retrieve a single subscription by its UUID
//	@Tags			subscriptions
//	@Produce		json
//	@Param			id	path		string	true	"Subscription UUID"
//	@Success		200	{object}	model.Subscription
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/subscriptions/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		slog.Error("handler.GetByID: internal error", "error", err)
		errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

// List godoc
//
//	@Summary		List subscriptions
//	@Description	Retrieve a list of subscriptions with optional filters
//	@Tags			subscriptions
//	@Produce		json
//	@Param			user_id			query		string	false	"Filter by user UUID"
//	@Param			service_name	query		string	false	"Filter by service name (case-insensitive)"
//	@Success		200				{array}		model.Subscription
//	@Failure		400				{object}	map[string]string
//	@Failure		500				{object}	map[string]string
//	@Router			/subscriptions [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	var userID, serviceName *string

	if v := r.URL.Query().Get("user_id"); v != "" {
		userID = &v
	}
	if v := r.URL.Query().Get("service_name"); v != "" {
		serviceName = &v
	}

	subs, err := h.svc.List(r.Context(), userID, serviceName)
	if err != nil {
		if errors.Is(err, service.ErrInvalid) {
			slog.Warn("handler.List: validation error", "error", err)
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("handler.List: internal error", "error", err)
		errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, subs)
}

// Update godoc
//
//	@Summary		Update a subscription
//	@Description	Fully update an existing subscription by ID
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Subscription UUID"
//	@Param			request	body		model.UpdateRequest	true	"Updated subscription data"
//	@Success		200		{object}	model.Subscription
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/subscriptions/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	var req model.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("handler.Update: bad request body", "error", err)
		errorResponse(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	sub, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		if errors.Is(err, service.ErrInvalid) {
			slog.Warn("handler.Update: validation error", "error", err)
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("handler.Update: internal error", "error", err)
		errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("handler.Update: subscription updated", "id", id)
	writeJSON(w, http.StatusOK, sub)
}

// Delete godoc
//
//	@Summary		Delete a subscription
//	@Description	Delete a subscription by ID
//	@Tags			subscriptions
//	@Produce		json
//	@Param			id	path	string	true	"Subscription UUID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/subscriptions/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			errorResponse(w, http.StatusNotFound, "subscription not found")
			return
		}
		slog.Error("handler.Delete: internal error", "error", err)
		errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("handler.Delete: subscription deleted", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

// TotalCost godoc
//
//	@Summary		Get total subscription cost
//	@Description	Calculate total cost of subscriptions over a period, with optional filters by user and service
//	@Tags			subscriptions
//	@Produce		json
//	@Param			start_period	query		string	true	"Period start in MM-YYYY format"
//	@Param			end_period		query		string	true	"Period end in MM-YYYY format"
//	@Param			user_id			query		string	false	"Filter by user UUID"
//	@Param			service_name	query		string	false	"Filter by service name (case-insensitive)"
//	@Success		200				{object}	model.TotalResponse
//	@Failure		400				{object}	map[string]string
//	@Failure		500				{object}	map[string]string
//	@Router			/subscriptions/total [get]
func (h *Handler) TotalCost(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	startPeriod := q.Get("start_period")
	endPeriod := q.Get("end_period")

	var userID, serviceName *string
	if v := q.Get("user_id"); v != "" {
		userID = &v
	}
	if v := q.Get("service_name"); v != "" {
		serviceName = &v
	}

	result, err := h.svc.TotalCost(r.Context(), startPeriod, endPeriod, userID, serviceName)
	if err != nil {
		if errors.Is(err, service.ErrInvalid) {
			slog.Warn("handler.TotalCost: validation error", "error", err)
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("handler.TotalCost: internal error", "error", err)
		errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// HealthCheck godoc
//
//	@Summary		Health check
//	@Description	Returns service health status
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
