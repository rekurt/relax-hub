package handler

import (
	"net/http"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type LoyaltyHandler struct {
	loyaltyService service.LoyaltyService
}

func NewLoyaltyHandler(loyaltyService service.LoyaltyService) *LoyaltyHandler {
	return &LoyaltyHandler{loyaltyService: loyaltyService}
}

type loyaltyAccountResponse struct {
	UserID     string                `json:"user_id"`
	Level      string                `json:"level"`
	Points     int64                 `json:"points"`
	TotalEarned int64               `json:"total_earned"`
	TotalSpent int64                 `json:"total_spent"`
	VisitCount int                   `json:"visit_count"`
	Privileges loyaltyPrivileges     `json:"privileges"`
	UpdatedAt  time.Time             `json:"updated_at"`
	CreatedAt  time.Time             `json:"created_at"`
}

type loyaltyPrivileges struct {
	PointMultiplier float64 `json:"point_multiplier"`
	DiscountPercent int     `json:"discount_percent"`
	NextLevel       *string `json:"next_level,omitempty"`
	VisitsToNext    *int    `json:"visits_to_next,omitempty"`
}

type loyaltyTransactionResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Type        string     `json:"type"`
	Amount      int64      `json:"amount"`
	BookingID   *string    `json:"booking_id,omitempty"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
}

type loyaltyLevelResponse struct {
	Level           string  `json:"level"`
	MinVisits       int     `json:"min_visits"`
	PointMultiplier float64 `json:"point_multiplier"`
	DiscountPercent int     `json:"discount_percent"`
}

func toLoyaltyAccountResponse(a *domain.LoyaltyAccount) loyaltyAccountResponse {
	levelInfo := domain.GetLoyaltyLevelInfo(a.Level)

	resp := loyaltyAccountResponse{
		UserID:      a.UserID.String(),
		Level:       string(a.Level),
		Points:      a.Points,
		TotalEarned: a.TotalEarned,
		TotalSpent:  a.TotalSpent,
		VisitCount:  a.VisitCount,
		Privileges: loyaltyPrivileges{
			PointMultiplier: levelInfo.PointMultiplier,
			DiscountPercent: levelInfo.DiscountPercent,
		},
		UpdatedAt: a.UpdatedAt,
		CreatedAt: a.CreatedAt,
	}

	// Calculate next level info
	allLevels := domain.GetAllLoyaltyLevels()
	for i, lvl := range allLevels {
		if lvl.Level == a.Level && i > 0 {
			nextLevel := string(allLevels[i-1].Level)
			visitsToNext := allLevels[i-1].MinVisits - a.VisitCount
			if visitsToNext < 0 {
				visitsToNext = 0
			}
			resp.Privileges.NextLevel = &nextLevel
			resp.Privileges.VisitsToNext = &visitsToNext
			break
		}
	}

	return resp
}

func toLoyaltyTransactionResponse(t *domain.LoyaltyTransaction) loyaltyTransactionResponse {
	resp := loyaltyTransactionResponse{
		ID:          t.ID.String(),
		UserID:      t.UserID.String(),
		Type:        string(t.Type),
		Amount:      t.Amount,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
	}
	if t.BookingID != nil {
		s := t.BookingID.String()
		resp.BookingID = &s
	}
	return resp
}

// GetAccount returns the loyalty account for the authenticated user.
func (h *LoyaltyHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	account, err := h.loyaltyService.GetAccount(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toLoyaltyAccountResponse(account))
}

// ListTransactions returns paginated transaction history for the authenticated user.
func (h *LoyaltyHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.loyaltyService.ListTransactions(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]loyaltyTransactionResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toLoyaltyTransactionResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// GetLevels returns information about all loyalty levels and their privileges.
func (h *LoyaltyHandler) GetLevels(w http.ResponseWriter, r *http.Request) {
	allLevels := domain.GetAllLoyaltyLevels()
	levels := make([]loyaltyLevelResponse, len(allLevels))
	// Reverse order so bronze is first
	for i, info := range allLevels {
		levels[len(allLevels)-1-i] = loyaltyLevelResponse{
			Level:           string(info.Level),
			MinVisits:       info.MinVisits,
			PointMultiplier: info.PointMultiplier,
			DiscountPercent: info.DiscountPercent,
		}
	}

	writeJSON(w, http.StatusOK, levels)
}
