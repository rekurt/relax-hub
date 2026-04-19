package handler

import (
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

const dateFormat = "2006-01-02"

type PricingHandler struct {
	pricingService      service.PricingService
	bathhouseService    service.BathhouseService
	smartPricingService service.SmartPricingService
	accessCheck         *service.AccessChecker
}

func NewPricingHandler(
	pricingService service.PricingService,
	bathhouseService service.BathhouseService,
	smartPricingService service.SmartPricingService,
	accessCheck *service.AccessChecker,
) *PricingHandler {
	return &PricingHandler{
		pricingService:      pricingService,
		bathhouseService:    bathhouseService,
		smartPricingService: smartPricingService,
		accessCheck:         accessCheck,
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

type pricingRuleRequest struct {
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

type seasonalTariffResponse struct {
	ID          string  `json:"id"`
	BathhouseID string  `json:"bathhouse_id"`
	Name        string  `json:"name"`
	DateFrom    string  `json:"date_from"`
	DateTo      string  `json:"date_to"`
	Multiplier  float64 `json:"multiplier"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type seasonalTariffRequest struct {
	Name       string  `json:"name"`
	DateFrom   string  `json:"date_from"`
	DateTo     string  `json:"date_to"`
	Multiplier float64 `json:"multiplier"`
	IsActive   bool    `json:"is_active"`
}

func toSeasonalTariffResponse(t *domain.SeasonalTariff) seasonalTariffResponse {
	return seasonalTariffResponse{
		ID:          t.ID.String(),
		BathhouseID: t.BathhouseID.String(),
		Name:        t.Name,
		DateFrom:    t.DateFrom.Format("2006-01-02"),
		DateTo:      t.DateTo.Format("2006-01-02"),
		Multiplier:  t.Multiplier,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
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

// CreateRule godoc
//
//	@Summary		Create pricing rule
//	@Description	Creates a new dynamic pricing rule for a bathhouse. Owner or representative only.
//	@Tags			pricing
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Bathhouse ID (UUID)"
//	@Param			body	body		pricingRuleRequest	true	"Pricing rule data"
//	@Success		201		{object}	APIResponse{data=pricingRuleResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/pricing-rules [post]
func (h *PricingHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req pricingRuleRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	// Validate rule type
	ruleType := domain.PricingRuleType(req.Type)
	if !ruleType.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid rule type")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        req.Name,
		Type:        ruleType,
		Multiplier:  req.Multiplier,
		DaysOfWeek:  req.DaysOfWeek,
		TimeFrom:    req.TimeFrom,
		TimeTo:      req.TimeTo,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		Priority:    req.Priority,
		IsActive:    req.IsActive,
	}

	createdRule, err := h.pricingService.CreateRule(r.Context(), userID, userRole, rule)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPricingRuleResponse(createdRule))
}

// ListRules godoc
//
//	@Summary		List pricing rules
//	@Description	Returns all pricing rules for a bathhouse. Owner or representative only.
//	@Tags			pricing
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]pricingRuleResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/pricing-rules [get]
func (h *PricingHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
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

// UpdateRule godoc
//
//	@Summary		Update pricing rule
//	@Description	Updates an existing pricing rule. Owner or representative only.
//	@Tags			pricing
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Rule ID (UUID)"
//	@Param			body	body		pricingRuleRequest	true	"Updated rule data"
//	@Success		200		{object}	APIResponse
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/pricing-rules/{id} [put]
func (h *PricingHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid rule id")
		return
	}

	var req pricingRuleRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	// Validate rule type
	ruleType := domain.PricingRuleType(req.Type)
	if !ruleType.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid rule type")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	rule := &domain.PricingRule{
		ID:         ruleID,
		Name:       req.Name,
		Type:       ruleType,
		Multiplier: req.Multiplier,
		DaysOfWeek: req.DaysOfWeek,
		TimeFrom:   req.TimeFrom,
		TimeTo:     req.TimeTo,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
		Priority:   req.Priority,
		IsActive:   req.IsActive,
	}

	if err := h.pricingService.UpdateRule(r.Context(), userID, userRole, rule); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "rule updated"})
}

// DeleteRule godoc
//
//	@Summary		Delete pricing rule
//	@Description	Deletes a pricing rule. Owner or representative only.
//	@Tags			pricing
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Rule ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/pricing-rules/{id} [delete]
func (h *PricingHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid rule id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.pricingService.DeleteRule(r.Context(), userID, userRole, ruleID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "rule deleted"})
}

// CalculatePrice godoc
//
//	@Summary		Calculate price
//	@Description	Calculates the price for a bathhouse booking with dynamic pricing rules applied. Public endpoint.
//	@Tags			pricing
//	@Produce		json
//	@Param			id		path		string	true	"Bathhouse ID (UUID)"
//	@Param			start	query		string	true	"Start time (RFC3339)"
//	@Param			end		query		string	true	"End time (RFC3339)"
//	@Success		200		{object}	APIResponse{data=priceCalculatorResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bathhouses/{id}/price-calculator [get]
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

	// Validate that start time is before end time
	if !startTime.Before(endTime) {
		writeError(w, http.StatusBadRequest, "invalid_input", "start must be before end")
		return
	}

	// Validate that times are hour-aligned (minute and second must be 0)
	if startTime.Minute() != 0 || startTime.Second() != 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "start time must be hour-aligned (minutes and seconds must be 0)")
		return
	}
	if endTime.Minute() != 0 || endTime.Second() != 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "end time must be hour-aligned (minutes and seconds must be 0)")
		return
	}

	// Get bathhouse to get base price
	bathhouse, err := h.bathhouseService.GetByID(r.Context(), bathhouseID)
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

	// Calculate base price for display - use ceil to handle potential floating-point rounding errors
	duration := endTime.Sub(startTime)
	hours := int64(math.Ceil(duration.Hours()))
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

// CreateSeasonalTariff godoc
//
//	@Summary		Create seasonal tariff
//	@Description	Creates a new seasonal tariff for a bathhouse. Owner or representative only.
//	@Tags			pricing
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Bathhouse ID (UUID)"
//	@Param			body	body		seasonalTariffRequest	true	"Seasonal tariff data"
//	@Success		201		{object}	APIResponse{data=seasonalTariffResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/seasonal-tariffs [post]
func (h *PricingHandler) CreateSeasonalTariff(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req seasonalTariffRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	dateFrom, err := time.Parse(dateFormat, req.DateFrom)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from format (use YYYY-MM-DD)")
		return
	}
	dateTo, err := time.Parse(dateFormat, req.DateTo)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to format (use YYYY-MM-DD)")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	tariff := &domain.SeasonalTariff{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		Name:        req.Name,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		Multiplier:  req.Multiplier,
		IsActive:    req.IsActive,
	}

	created, err := h.pricingService.CreateSeasonalTariff(r.Context(), userID, userRole, tariff)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toSeasonalTariffResponse(created))
}

// ListSeasonalTariffs godoc
//
//	@Summary		List seasonal tariffs
//	@Description	Returns all seasonal tariffs for a bathhouse. Owner or representative only.
//	@Tags			pricing
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]seasonalTariffResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/seasonal-tariffs [get]
func (h *PricingHandler) ListSeasonalTariffs(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	tariffs, err := h.pricingService.ListSeasonalTariffs(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	responses := make([]seasonalTariffResponse, len(tariffs))
	for i := range tariffs {
		responses[i] = toSeasonalTariffResponse(&tariffs[i])
	}

	writeJSON(w, http.StatusOK, responses)
}

// UpdateSeasonalTariff godoc
//
//	@Summary		Update seasonal tariff
//	@Description	Updates an existing seasonal tariff. Owner or representative only.
//	@Tags			pricing
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Tariff ID (UUID)"
//	@Param			body	body		seasonalTariffRequest	true	"Updated tariff data"
//	@Success		200		{object}	APIResponse
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/seasonal-tariffs/{id} [put]
func (h *PricingHandler) UpdateSeasonalTariff(w http.ResponseWriter, r *http.Request) {
	tariffID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid tariff id")
		return
	}

	var req seasonalTariffRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	dateFrom, err := time.Parse(dateFormat, req.DateFrom)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from format (use YYYY-MM-DD)")
		return
	}
	dateTo, err := time.Parse(dateFormat, req.DateTo)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to format (use YYYY-MM-DD)")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	tariff := &domain.SeasonalTariff{
		ID:         tariffID,
		Name:       req.Name,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Multiplier: req.Multiplier,
		IsActive:   req.IsActive,
	}

	if err := h.pricingService.UpdateSeasonalTariff(r.Context(), userID, userRole, tariff); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "tariff updated"})
}

// DeleteSeasonalTariff godoc
//
//	@Summary		Delete seasonal tariff
//	@Description	Deletes a seasonal tariff. Owner or representative only.
//	@Tags			pricing
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Tariff ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/seasonal-tariffs/{id} [delete]
func (h *PricingHandler) DeleteSeasonalTariff(w http.ResponseWriter, r *http.Request) {
	tariffID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid tariff id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.pricingService.DeleteSeasonalTariff(r.Context(), userID, userRole, tariffID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "tariff deleted"})
}

type priceRecommendationResponse struct {
	CurrentPrice        int64   `json:"current_price"`
	RecommendedPrice    int64   `json:"recommended_price"`
	Coefficient         float64 `json:"coefficient"`
	AvgAreaPrice        int64   `json:"avg_area_price"`
	OccupancyRate       float64 `json:"occupancy_rate"`
	DemandTrend         string  `json:"demand_trend"`
	RecommendationBasis string  `json:"recommendation_basis"`
}

// GetPriceRecommendation godoc
//
//	@Summary		Get smart pricing recommendation
//	@Description	Returns a pricing recommendation based on occupancy, area averages, and demand trends. Owner or representative only.
//	@Tags			pricing
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=priceRecommendationResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/price-recommendation [get]
func (h *PricingHandler) GetPriceRecommendation(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	rec, err := h.smartPricingService.GetRecommendation(r.Context(), userID, userRole, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, priceRecommendationResponse{
		CurrentPrice:        rec.CurrentPrice,
		RecommendedPrice:    rec.RecommendedPrice,
		Coefficient:         rec.Coefficient,
		AvgAreaPrice:        rec.AvgAreaPrice,
		OccupancyRate:       rec.OccupancyRate,
		DemandTrend:         rec.DemandTrend,
		RecommendationBasis: rec.RecommendationBasis,
	})
}
