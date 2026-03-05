package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/service"
)

type PricingHandler struct {
	pricingService  service.PricingService
	bathhouseRepo   repository.BathhouseRepository
	accessCheck     *service.AccessChecker
}

func NewPricingHandler(
	pricingService service.PricingService,
	bathhouseRepo repository.BathhouseRepository,
	accessCheck *service.AccessChecker,
) *PricingHandler {
	return &PricingHandler{
		pricingService:  pricingService,
		bathhouseRepo:   bathhouseRepo,
		accessCheck:     accessCheck,
	}
}

type pricingRuleResponse struct {
	ID          string     `json:"id"`
	BathhouseID string     `json:"bathhouse_id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Multiplier  float64    `json:"multiplier"`
	DaysOfWeek  []int      `json:"days_of_week,omitempty"`
	TimeFrom    *string    `json:"time_from,omitempty"`
	TimeTo      *string    `json:"time_to,omitempty"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	Priority    int        `json:"priority"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type createPricingRuleRequest struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Multiplier float64    `json:"multiplier"`
	DaysOfWeek []int      `json:"days_of_week,omitempty"`
	TimeFrom   *string    `json:"time_from,omitempty"`
	TimeTo     *string    `json:"time_to,omitempty"`
	DateFrom   *time.Time `json:"date_from,omitempty"`
	DateTo     *time.Time `json:"date_to,omitempty"`
	Priority   int        `json:"priority"`
	IsActive   bool       `json:"is_active"`
}

type updatePricingRuleRequest struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Multiplier float64    `json:"multiplier"`
	DaysOfWeek []int      `json:"days_of_week,omitempty"`
	TimeFrom   *string    `json:"time_from,omitempty"`
	TimeTo     *string    `json:"time_to,omitempty"`
	DateFrom   *time.Time `json:"date_from,omitempty"`
	DateTo     *time.Time `json:"date_to,omitempty"`
	Priority   int        `json:"priority"`
	IsActive   bool       `json:"is_active"`
}

type priceCalculatorResponse struct {
	BasePrice   int64 `json:"base_price"`
	FinalPrice  int64 `json:"final_price"`
	Hours       int64 `json:"hours"`
	PriceSaving int64 `json:"price_saving,omitempty"`
}

func toPricingRuleResponse(r *domain.PricingRule) pricingRuleResponse {
	return pricingRuleResponse{
		ID:          r.ID.String(),
		BathhouseID: r.BathhouseID.String(),
		Name:        r.Name,
		Type:        string(r.Type),
		Multiplier:  r.Multiplier,
		DaysOfWeek:  r.DaysOfWeek,
		TimeFrom:    r.TimeFrom,
		TimeTo:      r.TimeTo,
		DateFrom:    r.DateFrom,
		DateTo:      r.DateTo,
		Priority:    r.Priority,
		IsActive:    r.IsActive,
		CreatedAt:   r.CreatedAt,
	}
}

// CreateRule creates a new pricing rule
func (h *PricingHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req createPricingRuleRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	// Verify user has access to manage this bathhouse
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        req.Name,
		Type:        domain.PricingRuleType(req.Type),
		Multiplier:  req.Multiplier,
		DaysOfWeek:  req.DaysOfWeek,
		TimeFrom:    req.TimeFrom,
		TimeTo:      req.TimeTo,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		Priority:    req.Priority,
		IsActive:    req.IsActive,
		CreatedAt:   time.Now(),
	}

	createdRule, err := h.pricingService.CreateRule(r.Context(), userID, rule)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPricingRuleResponse(createdRule))
}

// ListRules returns all pricing rules for a bathhouse
func (h *PricingHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	// Verify user has access to manage this bathhouse
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	rules, err := h.pricingService.ListRules(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	responses := make([]pricingRuleResponse, len(rules))
	for i := range rules {
		responses[i] = toPricingRuleResponse(&rules[i])
	}

	writeJSON(w, http.StatusOK, responses)
}

// UpdateRule updates an existing pricing rule
func (h *PricingHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid rule id")
		return
	}

	var req updatePricingRuleRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	rule := &domain.PricingRule{
		ID:          ruleID,
		Name:        req.Name,
		Type:        domain.PricingRuleType(req.Type),
		Multiplier:  req.Multiplier,
		DaysOfWeek:  req.DaysOfWeek,
		TimeFrom:    req.TimeFrom,
		TimeTo:      req.TimeTo,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		Priority:    req.Priority,
		IsActive:    req.IsActive,
	}

	if err := h.pricingService.UpdateRule(r.Context(), userID, rule); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "rule updated"})
}

// DeleteRule deletes a pricing rule
func (h *PricingHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid rule id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	if err := h.pricingService.DeleteRule(r.Context(), userID, ruleID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "rule deleted"})
}

// CalculatePrice calculates price for a given bathhouse and time range
func (h *PricingHandler) CalculatePrice(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "start and end query parameters are required")
		return
	}

	// Parse ISO 8601 timestamps
	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid start time format (use RFC3339)")
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid end time format (use RFC3339)")
		return
	}

	// Get bathhouse to get base price
	bathhouse, err := h.bathhouseRepo.GetByID(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Calculate final price with rules
	finalPrice, err := h.pricingService.CalculatePrice(r.Context(), bathhouseID, bathhouse.PricePerHour, startTime, endTime)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Calculate base price (without multipliers)
	duration := endTime.Sub(startTime)
	hours := int64((duration.Hours() * 10) + 0.5) / 10 // Round to nearest 0.1
	if hours == 0 {
		hours = 1
	}
	basePrice := bathhouse.PricePerHour * hours

	resp := priceCalculatorResponse{
		BasePrice:  basePrice,
		FinalPrice: finalPrice,
		Hours:      hours,
	}

	if finalPrice < basePrice {
		resp.PriceSaving = basePrice - finalPrice
	}

	writeJSON(w, http.StatusOK, resp)
}
