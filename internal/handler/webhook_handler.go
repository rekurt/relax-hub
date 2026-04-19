package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type WebhookHandler struct {
	webhookService service.WebhookService
}

func NewWebhookHandler(webhookService service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

type createWebhookRequest struct {
	URL    string   `json:"url"`
	Secret string   `json:"secret"`
	Events []string `json:"events"`
}

type updateWebhookRequest struct {
	URL      string   `json:"url"`
	Secret   string   `json:"secret"`
	Events   []string `json:"events"`
	IsActive bool     `json:"is_active"`
}

type webhookResponse struct {
	ID        string   `json:"id"`
	OwnerID   string   `json:"owner_id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	IsActive  bool     `json:"is_active"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type webhookDeliveryResponse struct {
	ID           string  `json:"id"`
	WebhookID    string  `json:"webhook_id"`
	EventType    string  `json:"event_type"`
	Status       string  `json:"status"`
	HTTPStatus   int     `json:"http_status"`
	ErrorMessage string  `json:"error_message,omitempty"`
	AttemptCount int     `json:"attempt_count"`
	NextRetryAt  *string `json:"next_retry_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

func toWebhookResponse(w *domain.Webhook) webhookResponse {
	events := make([]string, len(w.Events))
	for i, e := range w.Events {
		events[i] = string(e)
	}
	return webhookResponse{
		ID:        w.ID.String(),
		OwnerID:   w.OwnerID.String(),
		URL:       w.URL,
		Events:    events,
		IsActive:  w.IsActive,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
}

func toWebhookDeliveryResponse(d *domain.WebhookDelivery) webhookDeliveryResponse {
	resp := webhookDeliveryResponse{
		ID:           d.ID.String(),
		WebhookID:    d.WebhookID.String(),
		EventType:    string(d.EventType),
		Status:       string(d.Status),
		HTTPStatus:   d.HTTPStatus,
		ErrorMessage: d.ErrorMessage,
		AttemptCount: d.AttemptCount,
		CreatedAt:    d.CreatedAt.Format(time.RFC3339),
	}
	if d.NextRetryAt != nil {
		s := d.NextRetryAt.Format(time.RFC3339)
		resp.NextRetryAt = &s
	}
	return resp
}

// CreateWebhook godoc
//
//	@Summary		Create a webhook
//	@Description	Create a new outgoing webhook for booking/payment events
//	@Tags			webhooks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createWebhookRequest	true	"Webhook data"
//	@Success		201		{object}	APIResponse{data=webhookResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/webhooks [post]
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var req createWebhookRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	events := make([]domain.WebhookEventType, len(req.Events))
	for i, e := range req.Events {
		events[i] = domain.WebhookEventType(e)
	}

	webhook := &domain.Webhook{
		ID:       uuid.New(),
		URL:      req.URL,
		Secret:   req.Secret,
		Events:   events,
		IsActive: true,
	}

	if err := h.webhookService.Create(r.Context(), userID, role, webhook); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toWebhookResponse(webhook))
}

// ListWebhooks godoc
//
//	@Summary		List webhooks
//	@Description	List all webhooks for the authenticated owner
//	@Tags			webhooks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]webhookResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/my/webhooks [get]
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	result, err := h.webhookService.List(r.Context(), userID, role, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []webhookResponse
	for _, wh := range result.Items {
		items = append(items, toWebhookResponse(&wh))
	}
	if items == nil {
		items = []webhookResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: int((result.TotalCount + int64(result.PageSize) - 1) / int64(result.PageSize)),
	})
}

// GetWebhook godoc
//
//	@Summary		Get a webhook
//	@Description	Get webhook details by ID
//	@Tags			webhooks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Webhook ID (UUID)"
//	@Success		200	{object}	APIResponse{data=webhookResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/webhooks/{id} [get]
func (h *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	webhookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid webhook ID")
		return
	}

	webhook, err := h.webhookService.GetByID(r.Context(), userID, role, webhookID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toWebhookResponse(webhook))
}

// UpdateWebhook godoc
//
//	@Summary		Update a webhook
//	@Description	Update an existing webhook
//	@Tags			webhooks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Webhook ID (UUID)"
//	@Param			body	body		updateWebhookRequest	true	"Webhook data"
//	@Success		200		{object}	APIResponse{data=webhookResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/webhooks/{id} [put]
func (h *WebhookHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	webhookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid webhook ID")
		return
	}

	var req updateWebhookRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	events := make([]domain.WebhookEventType, len(req.Events))
	for i, e := range req.Events {
		events[i] = domain.WebhookEventType(e)
	}

	webhook := &domain.Webhook{
		ID:       webhookID,
		URL:      req.URL,
		Secret:   req.Secret,
		Events:   events,
		IsActive: req.IsActive,
	}

	if err := h.webhookService.Update(r.Context(), userID, role, webhook); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toWebhookResponse(webhook))
}

// DeleteWebhook godoc
//
//	@Summary		Delete a webhook
//	@Description	Delete a webhook
//	@Tags			webhooks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Webhook ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/webhooks/{id} [delete]
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	webhookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid webhook ID")
		return
	}

	if err := h.webhookService.Delete(r.Context(), userID, role, webhookID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "webhook deleted"})
}

// ListDeliveries godoc
//
//	@Summary		List webhook deliveries
//	@Description	List delivery attempts for a webhook
//	@Tags			webhooks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Webhook ID (UUID)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]webhookDeliveryResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Failure		404			{object}	APIResponse{error=APIError}
//	@Router			/my/webhooks/{id}/deliveries [get]
func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	webhookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid webhook ID")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	result, err := h.webhookService.ListDeliveries(r.Context(), userID, role, webhookID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []webhookDeliveryResponse
	for _, d := range result.Items {
		items = append(items, toWebhookDeliveryResponse(&d))
	}
	if items == nil {
		items = []webhookDeliveryResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: int((result.TotalCount + int64(result.PageSize) - 1) / int64(result.PageSize)),
	})
}

// TestWebhook godoc
//
//	@Summary		Test a webhook
//	@Description	Send a test event to the webhook URL
//	@Tags			webhooks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Webhook ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/webhooks/{id}/test [post]
func (h *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	webhookID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid webhook ID")
		return
	}

	if err := h.webhookService.TestWebhook(r.Context(), userID, role, webhookID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "test webhook sent"})
}
