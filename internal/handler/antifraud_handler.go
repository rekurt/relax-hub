package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/antifraud"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type AntiFraudHandler struct {
	flagRepo   repository.FraudFlagRepository
	chatFilter antifraud.ChatFilter
}

func NewAntiFraudHandler(flagRepo repository.FraudFlagRepository, chatFilter antifraud.ChatFilter) *AntiFraudHandler {
	return &AntiFraudHandler{flagRepo: flagRepo, chatFilter: chatFilter}
}

type fraudFlagResponse struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	Rule       string  `json:"rule"`
	Severity   string  `json:"severity"`
	Status     string  `json:"status"`
	Action     string  `json:"action"`
	Details    interface{} `json:"details,omitempty"`
	CreatedAt  string  `json:"created_at"`
	ReviewedAt *string `json:"reviewed_at,omitempty"`
	ReviewedBy *string `json:"reviewed_by,omitempty"`
}

func toFraudFlagResponse(f *domain.FraudFlag) fraudFlagResponse {
	resp := fraudFlagResponse{
		ID:        f.ID.String(),
		UserID:    f.UserID.String(),
		Rule:      string(f.Rule),
		Severity:  string(f.Severity),
		Status:    string(f.Status),
		Action:    string(f.Action),
		CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if f.Details != nil {
		resp.Details = f.Details
	}
	if f.ReviewedAt != nil {
		t := f.ReviewedAt.Format("2006-01-02T15:04:05Z")
		resp.ReviewedAt = &t
	}
	if f.ReviewedBy != nil {
		s := f.ReviewedBy.String()
		resp.ReviewedBy = &s
	}
	return resp
}

func (h *AntiFraudHandler) ListFlags(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	filter := domain.FraudFlagFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.FraudFlagStatus(s)
		if status.IsValid() {
			filter.Status = &status
		}
	}
	if s := r.URL.Query().Get("rule"); s != "" {
		rule := domain.FraudRuleName(s)
		if rule.IsValid() {
			filter.Rule = &rule
		}
	}
	if s := r.URL.Query().Get("user_id"); s != "" {
		if uid, err := uuid.Parse(s); err == nil {
			filter.UserID = &uid
		}
	}

	result, err := h.flagRepo.ListPending(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]fraudFlagResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toFraudFlagResponse(&result.Items[i]))
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

type updateFraudFlagRequest struct {
	Status string `json:"status"`
}

func (h *AntiFraudHandler) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	flagID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid flag ID")
		return
	}

	var req updateFraudFlagRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	status := domain.FraudFlagStatus(req.Status)
	if !status.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_status", "invalid flag status")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.flagRepo.UpdateStatus(r.Context(), flagID, status, adminID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ListFilteredMessages godoc
// @Summary      List filtered chat messages
// @Description  Returns paginated list of chat messages that had contact information filtered
// @Tags         admin-antifraud
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int  false  "Page number"  default(1)
// @Param        page_size  query     int  false  "Page size"    default(20)
// @Success      200  {object}  APIResponse{data=[]antifraud.FilteredChatMessage,meta=Meta}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Router       /admin/chat/filtered [get]
func (h *AntiFraudHandler) ListFilteredMessages(w http.ResponseWriter, r *http.Request) {
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	items, total, err := h.chatFilter.ListFiltered(r.Context(), page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: totalPages,
	})
}
