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

type SubscriptionHandler struct {
	subService  service.SubscriptionService
	promoRepo   repository.PromotionRepository
	bhRepo      repository.BathhouseRepository
	accessCheck *service.AccessChecker
}

func NewSubscriptionHandler(
	subService service.SubscriptionService,
	promoRepo repository.PromotionRepository,
	bhRepo repository.BathhouseRepository,
	accessCheck *service.AccessChecker,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subService:  subService,
		promoRepo:   promoRepo,
		bhRepo:      bhRepo,
		accessCheck: accessCheck,
	}
}

type subscriptionResponse struct {
	ID           string    `json:"id"`
	BathhouseID  string    `json:"bathhouse_id"`
	OwnerID      string    `json:"owner_id"`
	Plan         string    `json:"plan"`
	Status       string    `json:"status"`
	StartDate    time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	AutoRenew    bool      `json:"auto_renew"`
	PriceKopecks int64     `json:"price_kopecks"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type promotionResponse struct {
	ID              string    `json:"id"`
	BathhouseID     string    `json:"bathhouse_id"`
	BudgetKopecks   int64     `json:"budget_kopecks"`
	SpentKopecks    int64     `json:"spent_kopecks"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	TargetCityID    *int64    `json:"target_city_id"`
	Status          string    `json:"status"`
	ImpressionCount int64     `json:"impression_count"`
	ClickCount      int64     `json:"click_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type subscribeRequest struct {
	Plan string `json:"plan"`
}

type createPromotionRequest struct {
	BudgetKopecks int64   `json:"budget_kopecks"`
	DurationDays  int     `json:"duration_days"`
	TargetCityID  *int64  `json:"target_city_id"`
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
		BudgetKopecks:   p.BudgetKopecks,
		SpentKopecks:    p.SpentKopecks,
		StartDate:       p.StartDate,
		EndDate:         p.EndDate,
		TargetCityID:    p.TargetCityID,
		Status:          string(p.Status),
		ImpressionCount: p.ImpressionCount,
		ClickCount:      p.ClickCount,
		CreatedAt:       p.CreatedAt,
	}
}

// Subscribe creates a new subscription for a bathhouse
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

	sub, err := h.subService.Subscribe(r.Context(), userID, bathhouseID, plan)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toSubscriptionResponse(sub))
}

// GetSubscription returns the active subscription for a bathhouse
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	// Verify user has access to this bathhouse
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, domain.RoleOwner, bathhouseID); err != nil {
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

// CancelSubscription cancels auto-renewal for a subscription
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	// Get the subscription first
	sub, err := h.subService.GetActive(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Cancel the subscription
	if err := h.subService.Cancel(r.Context(), userID, sub.ID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "subscription cancelled"})
}

// ListSubscriptions returns all subscriptions for the owner
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

// CreatePromotion creates a promotion campaign for a bathhouse
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

	if req.BudgetKopecks <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "budget must be positive")
		return
	}

	if req.DurationDays <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "duration_days must be positive")
		return
	}

	userID := middleware.GetUserID(r.Context())

	// Verify user has access to manage this bathhouse
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, domain.RoleOwner, bathhouseID); err != nil {
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
	existing, err := h.promoRepo.GetActiveBybathhouse(r.Context(), bathhouseID)
	if err != nil && err != domain.ErrNotFound {
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
		ID:            uuid.New(),
		BathhouseID:   bathhouseID,
		BudgetKopecks: req.BudgetKopecks,
		SpentKopecks:  0,
		StartDate:     now,
		EndDate:       endDate,
		TargetCityID:  req.TargetCityID,
		Status:        domain.PromotionActive,
		CreatedAt:     now,
	}

	if err := promo.Validate(); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.promoRepo.Create(r.Context(), promo); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPromotionResponse(promo))
}

// GetPromotion returns promotion statistics for a bathhouse
func (h *SubscriptionHandler) GetPromotion(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	// Verify user has access to this bathhouse
	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, domain.RoleOwner, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	promo, err := h.promoRepo.GetActiveBybathhouse(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPromotionResponse(promo))
}
