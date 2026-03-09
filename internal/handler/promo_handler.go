package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type PromoHandler struct {
	promoService service.PromoService
}

func NewPromoHandler(promoService service.PromoService) *PromoHandler {
	return &PromoHandler{promoService: promoService}
}

type createPromoRequest struct {
	Code       string    `json:"code"`
	Type       string    `json:"type"`
	Value      int64     `json:"value"`
	MaxUses    int       `json:"max_uses"`
	MinAmount  int64     `json:"min_amount"`
	ValidFrom  time.Time `json:"valid_from"`
	ValidUntil time.Time `json:"valid_until"`
}

type validatePromoRequest struct {
	Code        string `json:"code"`
	BathhouseID string `json:"bathhouse_id"`
	Amount      int64  `json:"amount"`
}

type promoResponse struct {
	ID          string     `json:"id"`
	Code        string     `json:"code"`
	Type        string     `json:"type"`
	Value       int64      `json:"value"`
	BathhouseID *string    `json:"bathhouse_id,omitempty"`
	CreatorID   string     `json:"creator_id"`
	MaxUses     int        `json:"max_uses"`
	CurrentUses int        `json:"current_uses"`
	MinAmount   int64      `json:"min_amount"`
	ValidFrom   time.Time  `json:"valid_from"`
	ValidUntil  time.Time  `json:"valid_until"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type validatePromoResponse struct {
	Code     string `json:"code"`
	Type     string `json:"type"`
	Value    int64  `json:"value"`
	Discount int64  `json:"discount"`
}

func toPromoResponse(p *domain.PromoCode) promoResponse {
	resp := promoResponse{
		ID:          p.ID.String(),
		Code:        p.Code,
		Type:        string(p.Type),
		Value:       p.Value,
		CreatorID:   p.CreatorID.String(),
		MaxUses:     p.MaxUses,
		CurrentUses: p.CurrentUses,
		MinAmount:   p.MinAmount,
		ValidFrom:   p.ValidFrom,
		ValidUntil:  p.ValidUntil,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
	}
	if p.BathhouseID != nil {
		s := p.BathhouseID.String()
		resp.BathhouseID = &s
	}
	return resp
}

func toPromoListResponse(promos []domain.PromoCode) []promoResponse {
	result := make([]promoResponse, len(promos))
	for i := range promos {
		result[i] = toPromoResponse(&promos[i])
	}
	return result
}

// CreateForBathhouse creates a promo code for a specific bathhouse.
func (h *PromoHandler) CreateForBathhouse(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseIDStr := chi.URLParam(r, "id")
	bathhouseID, err := uuid.Parse(bathhouseIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req createPromoRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	promo := &domain.PromoCode{
		Code:        req.Code,
		Type:        domain.PromoType(req.Type),
		Value:       req.Value,
		BathhouseID: &bathhouseID,
		MaxUses:     req.MaxUses,
		MinAmount:   req.MinAmount,
		ValidFrom:   req.ValidFrom,
		ValidUntil:  req.ValidUntil,
	}

	created, err := h.promoService.Create(r.Context(), userID, userRole, promo)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPromoResponse(created))
}

// ListByBathhouse returns promo codes for a specific bathhouse.
func (h *PromoHandler) ListByBathhouse(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseIDStr := chi.URLParam(r, "id")
	bathhouseID, err := uuid.Parse(bathhouseIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.promoService.ListByBathhouse(r.Context(), userID, userRole, bathhouseID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSONWithMeta(w, http.StatusOK, toPromoListResponse(result.Items), &Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// Deactivate deactivates a promo code.
func (h *PromoHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	promoIDStr := chi.URLParam(r, "id")
	promoID, err := uuid.Parse(promoIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid promo code id")
		return
	}

	if err := h.promoService.Deactivate(r.Context(), userID, userRole, promoID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"deactivated": true})
}

// Validate checks a promo code and returns the discount amount.
func (h *PromoHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req validatePromoRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	bathhouseID, err := uuid.Parse(req.BathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	promo, discount, err := h.promoService.Validate(r.Context(), req.Code, bathhouseID, req.Amount)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, validatePromoResponse{
		Code:     promo.Code,
		Type:     string(promo.Type),
		Value:    promo.Value,
		Discount: discount,
	})
}

// CreateGlobal creates a global promo code (admin only).
func (h *PromoHandler) CreateGlobal(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	var req createPromoRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	promo := &domain.PromoCode{
		Code:       req.Code,
		Type:       domain.PromoType(req.Type),
		Value:      req.Value,
		MaxUses:    req.MaxUses,
		MinAmount:  req.MinAmount,
		ValidFrom:  req.ValidFrom,
		ValidUntil: req.ValidUntil,
	}

	created, err := h.promoService.Create(r.Context(), userID, userRole, promo)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPromoResponse(created))
}
