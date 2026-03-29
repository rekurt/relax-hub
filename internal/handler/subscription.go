package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type SubscriptionHandler struct {
	subService   service.SubscriptionService
	promoService service.PromotionService
	accessCheck  *service.AccessChecker
}

func NewSubscriptionHandler(
	subService service.SubscriptionService,
	promoService service.PromotionService,
	accessCheck *service.AccessChecker,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subService:   subService,
		promoService: promoService,
		accessCheck:  accessCheck,
	}
}

type subscriptionResponse struct {
	ID           string     `json:"id"`
	BathhouseID  string     `json:"bathhouse_id"`
	OwnerID      string     `json:"owner_id"`
	Plan         string     `json:"plan"`
	Status       string     `json:"status"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	AutoRenew    bool       `json:"auto_renew"`
	PriceKopecks int64      `json:"price_kopecks"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type promotionResponse struct {
	ID              string    `json:"id"`
	BathhouseID     string    `json:"bathhouse_id"`
	DailyBidKopecks int64     `json:"daily_bid_kopecks"`
	BudgetKopecks   int64     `json:"budget_kopecks"`
	SpentKopecks    int64     `json:"spent_kopecks"`
	RemainingBudget int64     `json:"remaining_budget"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	TargetCityID    *int64    `json:"target_city_id"`
	Status          string    `json:"status"`
	ImpressionCount int64     `json:"impression_count"`
	ClickCount      int64     `json:"click_count"`
	CTR             float64   `json:"ctr"`
	CostPerClick    float64   `json:"cost_per_click"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type subscribeRequest struct {
	Plan string `json:"plan"`
}

type createPromotionRequest struct {
	DailyBidKopecks int64  `json:"daily_bid_kopecks"`
	BudgetKopecks   int64  `json:"budget_kopecks"`
	DurationDays    int    `json:"duration_days"`
	TargetCityID    *int64 `json:"target_city_id"`
}

type updatePromotionRequest struct {
	DailyBidKopecks *int64 `json:"daily_bid_kopecks,omitempty"`
	BudgetKopecks   *int64 `json:"budget_kopecks,omitempty"`
	EndDate         *string `json:"end_date,omitempty"`
}

func toSubscriptionResponse(s *domain.Subscription) subscriptionResponse {
	return subscriptionResponse{
		ID:           s.ID.String(),
		BathhouseID:  s.BathhouseID.String(),
		OwnerID:      s.OwnerID.String(),
		Plan:         string(s.Plan),
		Status:       string(s.Status),
		StartDate:    s.StartDate,
		EndDate:      s.EndDate,
		AutoRenew:    s.AutoRenew,
		PriceKopecks: s.PriceKopecks,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

func toPromotionResponse(p *domain.Promotion) promotionResponse {
	return promotionResponse{
		ID:              p.ID.String(),
		BathhouseID:     p.BathhouseID.String(),
		DailyBidKopecks: p.DailyBidKopecks,
		BudgetKopecks:   p.BudgetKopecks,
		SpentKopecks:    p.SpentKopecks,
		RemainingBudget: p.RemainingBudget(),
		StartDate:       p.StartDate,
		EndDate:         p.EndDate,
		TargetCityID:    p.TargetCityID,
		Status:          string(p.Status),
		ImpressionCount: p.ImpressionCount,
		ClickCount:      p.ClickCount,
		CTR:             p.CTR(),
		CostPerClick:    p.CostPerClick(),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// Subscribe godoc
//
//	@Summary		Create subscription
//	@Description	Creates a new subscription for a bathhouse. Owner or representative only.
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Bathhouse ID (UUID)"
//	@Param			body	body		subscribeRequest	true	"Subscription plan"
//	@Success		201		{object}	APIResponse{data=subscriptionResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/subscription [post]
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req subscribeRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	plan := domain.SubscriptionPlan(req.Plan)
	if !plan.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid plan")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	sub, err := h.subService.Subscribe(r.Context(), userID, userRole, bathhouseID, plan)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toSubscriptionResponse(sub))
}

// GetSubscription godoc
//
//	@Summary		Get active subscription
//	@Description	Returns the active subscription for a bathhouse. Owner or representative only.
//	@Tags			subscriptions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=subscriptionResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/subscription [get]
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	// Verify user has access to this bathhouse
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	sub, err := h.subService.GetActive(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toSubscriptionResponse(sub))
}

// CancelSubscription godoc
//
//	@Summary		Cancel subscription
//	@Description	Cancels auto-renewal for the active subscription of a bathhouse. Owner or representative only.
//	@Tags			subscriptions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/subscription [delete]
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	// Get the subscription first
	sub, err := h.subService.GetActive(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Cancel the subscription
	if err := h.subService.Cancel(r.Context(), userID, userRole, sub.ID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "subscription cancelled"})
}

// ListSubscriptions godoc
//
//	@Summary		List my subscriptions
//	@Description	Returns a paginated list of all subscriptions for the authenticated owner.
//	@Tags			subscriptions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]subscriptionResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.subService.ListByOwner(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]subscriptionResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toSubscriptionResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// CreatePromotion godoc
//
//	@Summary		Create promotion
//	@Description	Creates a promotion campaign for a bathhouse. Requires an active Promoted subscription. Owner or representative only.
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Bathhouse ID (UUID)"
//	@Param			body	body		createPromotionRequest	true	"Promotion parameters"
//	@Success		201		{object}	APIResponse{data=promotionResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/promotion [post]
func (h *SubscriptionHandler) CreatePromotion(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req createPromotionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.DailyBidKopecks < domain.MinDailyBidKopecks {
		writeError(w, http.StatusBadRequest, "invalid_input", "daily_bid_kopecks must be at least 5000 (50 rubles)")
		return
	}

	if req.BudgetKopecks <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "budget must be positive")
		return
	}

	if req.BudgetKopecks < req.DailyBidKopecks {
		writeError(w, http.StatusBadRequest, "invalid_input", "budget must be at least equal to daily bid")
		return
	}

	if req.DurationDays <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "duration_days must be positive")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	// Verify user has access to manage this bathhouse
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	// Verify bathhouse has a Promoted subscription
	sub, err := h.subService.GetActive(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if sub.Plan != domain.PlanPromoted {
		writeError(w, http.StatusBadRequest, "no_promoted_subscription", "bathhouse must have an active Promoted subscription to create a promotion")
		return
	}

	// Check if there's already an active promotion
	existing, err := h.promoService.GetActiveBybathhouse(r.Context(), bathhouseID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		handleServiceError(w, err)
		return
	}
	if existing != nil && existing.Status == domain.PromotionActive {
		writeError(w, http.StatusConflict, "promotion_already_active", "bathhouse already has an active promotion")
		return
	}

	now := time.Now()
	endDate := now.AddDate(0, 0, req.DurationDays)

	promo := &domain.Promotion{
		ID:              uuid.New(),
		BathhouseID:     bathhouseID,
		DailyBidKopecks: req.DailyBidKopecks,
		BudgetKopecks:   req.BudgetKopecks,
		SpentKopecks:    0,
		StartDate:       now,
		EndDate:         endDate,
		TargetCityID:    req.TargetCityID,
		Status:          domain.PromotionActive,
		CreatedAt:       now,
	}

	if err := promo.Validate(); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.promoService.Create(r.Context(), promo); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPromotionResponse(promo))
}

// GetPromotion godoc
//
//	@Summary		Get promotion
//	@Description	Returns promotion statistics for a bathhouse. Owner or representative only.
//	@Tags			subscriptions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=promotionResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/promotion [get]
func (h *SubscriptionHandler) GetPromotion(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	// Verify user has access to this bathhouse
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	promo, err := h.promoService.GetActiveBybathhouse(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPromotionResponse(promo))
}

// UpdatePromotion godoc
//
//	@Summary		Update promotion campaign
//	@Description	Updates a promotion campaign parameters (daily bid, budget, end date). Owner or representative only.
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Promotion ID (UUID)"
//	@Param			body	body		updatePromotionRequest	true	"Update parameters"
//	@Success		200		{object}	APIResponse{data=promotionResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/promotions/{id} [put]
func (h *SubscriptionHandler) UpdatePromotion(w http.ResponseWriter, r *http.Request) {
	promoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid promotion id")
		return
	}

	var req updatePromotionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	promo, err := h.promoService.GetByID(r.Context(), promoID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, promo.BathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.DailyBidKopecks != nil {
		if *req.DailyBidKopecks < domain.MinDailyBidKopecks {
			writeError(w, http.StatusBadRequest, "invalid_input", "daily_bid_kopecks must be at least 5000 (50 rubles)")
			return
		}
		promo.DailyBidKopecks = *req.DailyBidKopecks
	}

	if req.BudgetKopecks != nil {
		if *req.BudgetKopecks <= promo.SpentKopecks {
			writeError(w, http.StatusBadRequest, "invalid_input", "new budget must be greater than already spent amount")
			return
		}
		promo.BudgetKopecks = *req.BudgetKopecks
	}

	if req.EndDate != nil {
		endDate, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid end_date format (RFC3339)")
			return
		}
		if !endDate.After(time.Now()) {
			writeError(w, http.StatusBadRequest, "invalid_input", "end_date must be in the future")
			return
		}
		promo.EndDate = endDate
	}

	promo.UpdatedAt = time.Now()
	if err := h.promoService.Update(r.Context(), promo); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPromotionResponse(promo))
}

// PausePromotion godoc
//
//	@Summary		Pause promotion campaign
//	@Description	Pauses an active promotion campaign. Owner or representative only.
//	@Tags			subscriptions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Promotion ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/promotions/{id}/pause [post]
func (h *SubscriptionHandler) PausePromotion(w http.ResponseWriter, r *http.Request) {
	promoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid promotion id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	promo, err := h.promoService.GetByID(r.Context(), promoID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, promo.BathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.promoService.Pause(r.Context(), promoID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "promotion paused"})
}

// ResumePromotion godoc
//
//	@Summary		Resume promotion campaign
//	@Description	Resumes a paused promotion campaign. Owner or representative only.
//	@Tags			subscriptions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Promotion ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Failure		409	{object}	APIResponse{error=APIError}
//	@Router			/my/promotions/{id}/resume [post]
func (h *SubscriptionHandler) ResumePromotion(w http.ResponseWriter, r *http.Request) {
	promoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid promotion id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	promo, err := h.promoService.GetByID(r.Context(), promoID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, promo.BathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.promoService.Resume(r.Context(), promoID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "promotion resumed"})
}

// ListPromotions godoc
//
//	@Summary		List promotion campaigns
//	@Description	Returns a paginated list of all promotion campaigns for a bathhouse. Owner or representative only.
//	@Tags			subscriptions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Bathhouse ID (UUID)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]promotionResponse,meta=Meta}
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/promotions [get]
func (h *SubscriptionHandler) ListPromotions(w http.ResponseWriter, r *http.Request) {
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

	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.promoService.ListByBathhouse(r.Context(), bathhouseID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]promotionResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toPromotionResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
