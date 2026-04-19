package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/service"
)

type FAQHandler struct {
	faqService service.FAQBotService
}

func NewFAQHandler(faqService service.FAQBotService) *FAQHandler {
	return &FAQHandler{faqService: faqService}
}

type createFAQRequest struct {
	Category  string   `json:"category"`
	Question  string   `json:"question"`
	Answer    string   `json:"answer"`
	Keywords  []string `json:"keywords"`
	SortOrder int      `json:"sort_order"`
}

type updateFAQRequest struct {
	Category  string   `json:"category"`
	Question  string   `json:"question"`
	Answer    string   `json:"answer"`
	Keywords  []string `json:"keywords"`
	SortOrder int      `json:"sort_order"`
	Active    bool     `json:"active"`
}

type faqResponse struct {
	ID        string   `json:"id"`
	Category  string   `json:"category"`
	Question  string   `json:"question"`
	Answer    string   `json:"answer"`
	Keywords  []string `json:"keywords"`
	SortOrder int      `json:"sort_order"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type faqMatchResponse struct {
	FAQ   faqResponse `json:"faq"`
	Score float64     `json:"score"`
}

type matchFAQRequest struct {
	Query string `json:"query"`
}

func toFAQResponse(f *domain.FAQ) faqResponse {
	return faqResponse{
		ID:        f.ID.String(),
		Category:  string(f.Category),
		Question:  f.Question,
		Answer:    f.Answer,
		Keywords:  f.Keywords,
		SortOrder: f.SortOrder,
		Active:    f.Active,
		CreatedAt: f.CreatedAt.Format(time.RFC3339),
		UpdatedAt: f.UpdatedAt.Format(time.RFC3339),
	}
}

func toFAQListResponse(items []domain.FAQ) []faqResponse {
	result := make([]faqResponse, len(items))
	for i := range items {
		result[i] = toFAQResponse(&items[i])
	}
	return result
}

// PublicListFAQ godoc
//
//	@Summary		List public FAQ entries
//	@Description	Returns active FAQ entries for the public site
//	@Tags			faq
//	@Produce		json
//	@Param			category	query		string	false	"Filter by category"
//	@Success		200			{object}	APIResponse{data=[]faqResponse,meta=Meta}
//	@Router			/faq [get]
func (h *FAQHandler) PublicListFAQ(w http.ResponseWriter, r *http.Request) {
	active := true
	filter := domain.FAQFilter{
		Page:     1,
		PageSize: 100,
		Active:   &active,
	}

	if cat := r.URL.Query().Get("category"); cat != "" {
		category := domain.FAQCategory(cat)
		filter.Category = &category
	}

	result, err := h.faqService.ListFAQ(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSONWithMeta(w, http.StatusOK, toFAQListResponse(result.Items), &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// AdminCreateFAQ godoc
//
//	@Summary		Create FAQ entry
//	@Description	Creates a new FAQ entry. Admin only.
//	@Tags			faq
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createFAQRequest	true	"FAQ data"
//	@Success		201		{object}	APIResponse{data=faqResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Router			/admin/faq [post]
func (h *FAQHandler) AdminCreateFAQ(w http.ResponseWriter, r *http.Request) {
	var req createFAQRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	faq := &domain.FAQ{
		Category:  domain.FAQCategory(req.Category),
		Question:  req.Question,
		Answer:    req.Answer,
		Keywords:  req.Keywords,
		SortOrder: req.SortOrder,
	}

	created, err := h.faqService.CreateFAQ(r.Context(), faq)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toFAQResponse(created))
}

// AdminUpdateFAQ godoc
//
//	@Summary		Update FAQ entry
//	@Description	Updates an existing FAQ entry. Admin only.
//	@Tags			faq
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"FAQ ID"
//	@Param			body	body		updateFAQRequest	true	"FAQ data"
//	@Success		200		{object}	APIResponse{data=faqResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/faq/{id} [put]
func (h *FAQHandler) AdminUpdateFAQ(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid FAQ ID")
		return
	}

	var req updateFAQRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	faq := &domain.FAQ{
		ID:        id,
		Category:  domain.FAQCategory(req.Category),
		Question:  req.Question,
		Answer:    req.Answer,
		Keywords:  req.Keywords,
		SortOrder: req.SortOrder,
		Active:    req.Active,
	}

	updated, err := h.faqService.UpdateFAQ(r.Context(), faq)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toFAQResponse(updated))
}

// AdminDeleteFAQ godoc
//
//	@Summary		Delete FAQ entry
//	@Description	Deletes an FAQ entry. Admin only.
//	@Tags			faq
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"FAQ ID"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/faq/{id} [delete]
func (h *FAQHandler) AdminDeleteFAQ(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid FAQ ID")
		return
	}

	if err := h.faqService.DeleteFAQ(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "FAQ deleted"})
}

// AdminGetFAQ godoc
//
//	@Summary		Get FAQ entry
//	@Description	Returns a single FAQ entry by ID. Admin only.
//	@Tags			faq
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"FAQ ID"
//	@Success		200	{object}	APIResponse{data=faqResponse}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/faq/{id} [get]
func (h *FAQHandler) AdminGetFAQ(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid FAQ ID")
		return
	}

	faq, err := h.faqService.GetFAQ(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toFAQResponse(faq))
}

// AdminListFAQ godoc
//
//	@Summary		List FAQ entries
//	@Description	Returns a paginated list of FAQ entries with optional filters. Admin only.
//	@Tags			faq
//	@Produce		json
//	@Security		BearerAuth
//	@Param			category	query		string	false	"Filter by category"
//	@Param			active		query		bool	false	"Filter by active status"
//	@Param			page		query		int		false	"Page number"
//	@Param			page_size	query		int		false	"Items per page"
//	@Success		200			{object}	APIResponse{data=[]faqResponse}
//	@Router			/admin/faq [get]
func (h *FAQHandler) AdminListFAQ(w http.ResponseWriter, r *http.Request) {
	filter := domain.FAQFilter{
		Page:     1,
		PageSize: 20,
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			filter.Page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil {
			filter.PageSize = v
		}
	}
	if cat := r.URL.Query().Get("category"); cat != "" {
		c := domain.FAQCategory(cat)
		filter.Category = &c
	}
	if act := r.URL.Query().Get("active"); act != "" {
		if v, err := strconv.ParseBool(act); err == nil {
			filter.Active = &v
		}
	}

	result, err := h.faqService.ListFAQ(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSONWithMeta(w, http.StatusOK, toFAQListResponse(result.Items), &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// MatchFAQ godoc
//
//	@Summary		Match FAQ entries
//	@Description	Search FAQ entries by query text. Returns top 3 matching FAQs. Used by support chat widget before ticket creation.
//	@Tags			faq
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		matchFAQRequest	true	"Search query"
//	@Success		200		{object}	APIResponse{data=[]faqMatchResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Router			/my/support/faq-match [post]
func (h *FAQHandler) MatchFAQ(w http.ResponseWriter, r *http.Request) {
	var req matchFAQRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	matches, err := h.faqService.MatchFAQ(r.Context(), req.Query)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := make([]faqMatchResponse, len(matches))
	for i, m := range matches {
		resp[i] = faqMatchResponse{
			FAQ:   toFAQResponse(&m.FAQ),
			Score: m.Score,
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// SeedFAQ godoc
//
//	@Summary		Seed default FAQ entries
//	@Description	Seeds the FAQ table with default entries if empty. Admin only.
//	@Tags			faq
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Router			/admin/faq/seed [post]
func (h *FAQHandler) SeedFAQ(w http.ResponseWriter, r *http.Request) {
	if err := h.faqService.SeedDefaultFAQ(r.Context()); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "FAQ entries seeded"})
}
