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

type BroadcastHandler struct {
	broadcastService service.BroadcastService
}

func NewBroadcastHandler(broadcastService service.BroadcastService) *BroadcastHandler {
	return &BroadcastHandler{broadcastService: broadcastService}
}

type createBroadcastRequest struct {
	Segment     string   `json:"segment"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	ImageURL    string   `json:"image_url"`
	PromoCodeID *string  `json:"promo_code_id"`
	Channels    []string `json:"channels"`
}

type broadcastResponse struct {
	ID          string   `json:"id"`
	OwnerID     string   `json:"owner_id"`
	Segment     string   `json:"segment"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	ImageURL    string   `json:"image_url"`
	PromoCodeID *string  `json:"promo_code_id,omitempty"`
	Channels    []string `json:"channels"`
	Status      string   `json:"status"`
	Delivered   int64    `json:"delivered"`
	Read        int64    `json:"read"`
	Clicked     int64    `json:"clicked"`
	SentAt      *string  `json:"sent_at,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func toBroadcastResponse(b *domain.Broadcast) broadcastResponse {
	channels := make([]string, len(b.Channels))
	for i, ch := range b.Channels {
		channels[i] = string(ch)
	}
	resp := broadcastResponse{
		ID:        b.ID.String(),
		OwnerID:   b.OwnerID.String(),
		Segment:   string(b.Segment),
		Title:     b.Title,
		Body:      b.Body,
		ImageURL:  b.ImageURL,
		Channels:  channels,
		Status:    string(b.Status),
		Delivered: b.Delivered,
		Read:      b.Read,
		Clicked:   b.Clicked,
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
		UpdatedAt: b.UpdatedAt.Format(time.RFC3339),
	}
	if b.PromoCodeID != nil {
		s := b.PromoCodeID.String()
		resp.PromoCodeID = &s
	}
	if b.SentAt != nil {
		s := b.SentAt.Format(time.RFC3339)
		resp.SentAt = &s
	}
	return resp
}

// CreateBroadcast godoc
//
//	@Summary		Create a broadcast
//	@Description	Create a new broadcast message for a guest segment
//	@Tags			crm
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createBroadcastRequest	true	"Broadcast data"
//	@Success		201		{object}	APIResponse{data=broadcastResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/crm/broadcasts [post]
func (h *BroadcastHandler) CreateBroadcast(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var req createBroadcastRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	channels := make([]domain.BroadcastChannel, len(req.Channels))
	for i, ch := range req.Channels {
		channels[i] = domain.BroadcastChannel(ch)
	}

	broadcast := &domain.Broadcast{
		ID:       uuid.New(),
		Segment:  domain.GuestSegmentSlug(req.Segment),
		Title:    req.Title,
		Body:     req.Body,
		ImageURL: req.ImageURL,
		Channels: channels,
	}

	if req.PromoCodeID != nil && *req.PromoCodeID != "" {
		id, err := uuid.Parse(*req.PromoCodeID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_promo_code_id", "invalid promo code ID")
			return
		}
		broadcast.PromoCodeID = &id
	}

	if err := h.broadcastService.Create(r.Context(), userID, role, broadcast); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toBroadcastResponse(broadcast))
}

// ListBroadcasts godoc
//
//	@Summary		List broadcasts
//	@Description	Get paginated list of broadcasts for the owner
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status		query		string	false	"Filter by status (draft, sending, sent, failed)"
//	@Param			page		query		int		false	"Page number"
//	@Param			page_size	query		int		false	"Page size"
//	@Success		200			{object}	APIResponse{data=[]broadcastResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/my/crm/broadcasts [get]
func (h *BroadcastHandler) ListBroadcasts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	filter := domain.BroadcastFilter{}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.BroadcastStatus(status)
		filter.Status = &s
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}
	filter.Page = page
	filter.PageSize = pageSize

	result, err := h.broadcastService.ListBroadcasts(r.Context(), userID, role, filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []broadcastResponse
	for _, b := range result.Items {
		items = append(items, toBroadcastResponse(&b))
	}
	if items == nil {
		items = []broadcastResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// SendBroadcast godoc
//
//	@Summary		Send a broadcast
//	@Description	Send a draft broadcast to the target segment
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Broadcast ID (UUID)"
//	@Success		200	{object}	APIResponse{data=broadcastResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Failure		429	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/broadcasts/{id}/send [post]
func (h *BroadcastHandler) SendBroadcast(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	broadcastID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid broadcast ID")
		return
	}

	if err := h.broadcastService.Send(r.Context(), userID, role, broadcastID); err != nil {
		handleServiceError(w, err)
		return
	}

	// Fetch updated broadcast to return
	broadcast, err := h.broadcastService.GetBroadcast(r.Context(), userID, role, broadcastID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBroadcastResponse(broadcast))
}

// GetBroadcast godoc
//
//	@Summary		Get broadcast details
//	@Description	Get a specific broadcast by ID
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Broadcast ID (UUID)"
//	@Success		200	{object}	APIResponse{data=broadcastResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/broadcasts/{id} [get]
func (h *BroadcastHandler) GetBroadcast(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	broadcastID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid broadcast ID")
		return
	}

	broadcast, err := h.broadcastService.GetBroadcast(r.Context(), userID, role, broadcastID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBroadcastResponse(broadcast))
}
