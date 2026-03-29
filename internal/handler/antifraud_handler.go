package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/antifraud"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type AntiFraudHandler struct {
	flagRepo     repository.FraudFlagRepository
	chatFilter   antifraud.ChatFilter
	stoplistRepo repository.StoplistRepository
}

func NewAntiFraudHandler(flagRepo repository.FraudFlagRepository, chatFilter antifraud.ChatFilter, stoplistRepo repository.StoplistRepository) *AntiFraudHandler {
	return &AntiFraudHandler{flagRepo: flagRepo, chatFilter: chatFilter, stoplistRepo: stoplistRepo}
}

type fraudFlagResponse struct {
	ID         string      `json:"id"`
	UserID     string      `json:"user_id"`
	Rule       string      `json:"rule"`
	Severity   string      `json:"severity"`
	Status     string      `json:"status"`
	Action     string      `json:"action"`
	Details    interface{} `json:"details,omitempty"`
	CreatedAt  string      `json:"created_at"`
	ReviewedAt *string     `json:"reviewed_at,omitempty"`
	ReviewedBy *string     `json:"reviewed_by,omitempty"`
}

func toFraudFlagResponse(f *domain.FraudFlag) fraudFlagResponse {
	resp := fraudFlagResponse{
		ID:        f.ID.String(),
		UserID:    f.UserID.String(),
		Rule:      string(f.Rule),
		Severity:  string(f.Severity),
		Status:    string(f.Status),
		Action:    string(f.Action),
		CreatedAt: f.CreatedAt.Format(time.RFC3339),
	}
	if f.Details != nil {
		resp.Details = f.Details
	}
	if f.ReviewedAt != nil {
		t := f.ReviewedAt.Format(time.RFC3339)
		resp.ReviewedAt = &t
	}
	if f.ReviewedBy != nil {
		s := f.ReviewedBy.String()
		resp.ReviewedBy = &s
	}
	return resp
}

// ListFlags godoc
//
//	@Summary		List fraud flags
//	@Description	Returns paginated list of fraud flags with optional filters by status, rule, and user
//	@Tags			admin-antifraud
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status		query		string	false	"Filter by flag status (pending, reviewed, dismissed)"
//	@Param			rule		query		string	false	"Filter by fraud rule name"
//	@Param			user_id		query		string	false	"Filter by user ID (UUID)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]fraudFlagResponse,meta=Meta}
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/admin/antifraud/flags [get]
func (h *AntiFraudHandler) ListFlags(w http.ResponseWriter, r *http.Request) {
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	filter := domain.FraudFlagFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.FraudFlagStatus(s)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_status", "invalid fraud flag status")
			return
		}
		filter.Status = &status
	}
	if s := r.URL.Query().Get("rule"); s != "" {
		rule := domain.FraudRuleName(s)
		if !rule.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_rule", "invalid fraud rule name")
			return
		}
		filter.Rule = &rule
	}
	if s := r.URL.Query().Get("user_id"); s != "" {
		uid, err := uuid.Parse(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_user_id", "invalid user ID format")
			return
		}
		filter.UserID = &uid
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

// UpdateFlag godoc
//
//	@Summary		Update fraud flag status
//	@Description	Update the status of a fraud flag (reviewed or dismissed)
//	@Tags			admin-antifraud
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Fraud flag ID (UUID)"
//	@Param			body	body		updateFraudFlagRequest	true	"New status"
//	@Success		200		{object}	APIResponse{data=object}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/antifraud/flags/{id} [patch]
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
	if status != domain.FraudFlagStatusReviewed && status != domain.FraudFlagStatusDismissed {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be 'reviewed' or 'dismissed'")
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
//
//	@Summary		List filtered chat messages
//	@Description	Returns paginated list of chat messages that had contact information filtered
//	@Tags			admin-antifraud
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]antifraud.FilteredChatMessage,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/admin/chat/filtered [get]
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

type stoplistEntryResponse struct {
	ID             string  `json:"id"`
	Phone          string  `json:"phone,omitempty"`
	Email          string  `json:"email,omitempty"`
	INN            string  `json:"inn,omitempty"`
	BankCardNumber string  `json:"bank_card_number,omitempty"`
	Reason         string  `json:"reason"`
	BlockedAt      string  `json:"blocked_at"`
	CreatedBy      *string `json:"created_by,omitempty"`
}

func toStoplistResponse(e *domain.StoplistEntry) stoplistEntryResponse {
	resp := stoplistEntryResponse{
		ID:             e.ID.String(),
		Phone:          e.Phone,
		Email:          e.Email,
		INN:            e.INN,
		BankCardNumber: e.BankCardNumber,
		Reason:         e.Reason,
		BlockedAt:      e.BlockedAt.Format(time.RFC3339),
	}
	if e.CreatedBy != nil {
		s := e.CreatedBy.String()
		resp.CreatedBy = &s
	}
	return resp
}

type createStoplistRequest struct {
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	INN            string `json:"inn"`
	BankCardNumber string `json:"bank_card_number"`
	Reason         string `json:"reason"`
}

// ListStoplist godoc
//
//	@Summary		List antifraud stoplist entries
//	@Description	Returns paginated list of stoplist entries
//	@Tags			admin-antifraud
//	@Produce		json
//	@Security		BearerAuth
//	@Param			phone		query		string	false	"Filter by phone"
//	@Param			email		query		string	false	"Filter by email"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]stoplistEntryResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/admin/antifraud/stoplist [get]
func (h *AntiFraudHandler) ListStoplist(w http.ResponseWriter, r *http.Request) {
	filter := domain.StoplistFilter{
		Phone:    r.URL.Query().Get("phone"),
		Email:    r.URL.Query().Get("email"),
		Page:     getPage(r.URL.Query().Get("page")),
		PageSize: getPageSize(r.URL.Query().Get("page_size"), 20),
	}

	result, err := h.stoplistRepo.List(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]stoplistEntryResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toStoplistResponse(&result.Items[i]))
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// CreateStoplistEntry godoc
//
//	@Summary		Add entry to antifraud stoplist
//	@Description	Block phone, email, INN or bank card from creating listings
//	@Tags			admin-antifraud
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createStoplistRequest	true	"Stoplist entry"
//	@Success		201		{object}	APIResponse{data=stoplistEntryResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/antifraud/stoplist [post]
func (h *AntiFraudHandler) CreateStoplistEntry(w http.ResponseWriter, r *http.Request) {
	var req createStoplistRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}
	if req.Phone == "" && req.Email == "" && req.INN == "" && req.BankCardNumber == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "at least one identifier (phone, email, inn, bank_card_number) is required")
		return
	}

	adminID := middleware.GetUserID(r.Context())
	entry := &domain.StoplistEntry{
		Phone:          req.Phone,
		Email:          req.Email,
		INN:            req.INN,
		BankCardNumber: req.BankCardNumber,
		Reason:         req.Reason,
		CreatedBy:      &adminID,
	}

	if err := h.stoplistRepo.Create(r.Context(), entry); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toStoplistResponse(entry))
}

// DeleteStoplistEntry godoc
//
//	@Summary		Remove entry from antifraud stoplist
//	@Description	Delete a stoplist entry by ID
//	@Tags			admin-antifraud
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Stoplist entry ID (UUID)"
//	@Success		200	{object}	APIResponse{data=object}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/antifraud/stoplist/{id} [delete]
func (h *AntiFraudHandler) DeleteStoplistEntry(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid stoplist entry ID")
		return
	}

	if err := h.stoplistRepo.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
