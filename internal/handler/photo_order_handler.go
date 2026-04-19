package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

// PhotoOrderHandler handles professional photography order endpoints.
type PhotoOrderHandler struct {
	svc service.PhotoOrderService
}

func NewPhotoOrderHandler(svc service.PhotoOrderService) *PhotoOrderHandler {
	return &PhotoOrderHandler{svc: svc}
}

type createPhotoOrderRequest struct {
	Region string `json:"region"`
	Notes  string `json:"notes"`
}

type photoOrderResponse struct {
	ID               string  `json:"id"`
	OwnerID          string  `json:"owner_id"`
	BathhouseID      string  `json:"bathhouse_id"`
	Region           string  `json:"region"`
	Status           string  `json:"status"`
	PhotographerName string  `json:"photographer_name,omitempty"`
	Price            int64   `json:"price"`
	ScheduledAt      *string `json:"scheduled_at,omitempty"`
	Notes            string  `json:"notes,omitempty"`
	AdminNotes       string  `json:"admin_notes,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func toPhotoOrderResponse(o *domain.PhotoOrder) photoOrderResponse {
	resp := photoOrderResponse{
		ID:               o.ID.String(),
		OwnerID:          o.OwnerID.String(),
		BathhouseID:      o.BathhouseID.String(),
		Region:           o.Region,
		Status:           string(o.Status),
		PhotographerName: o.PhotographerName,
		Price:            o.Price,
		Notes:            o.Notes,
		AdminNotes:       o.AdminNotes,
		CreatedAt:        o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        o.UpdatedAt.Format(time.RFC3339),
	}
	if o.ScheduledAt != nil {
		s := o.ScheduledAt.Format(time.RFC3339)
		resp.ScheduledAt = &s
	}
	return resp
}

// Create godoc
//
//	@Summary		Request professional photography
//	@Tags			photo-orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Bathhouse ID"
//	@Param			body	body		createPhotoOrderRequest		true	"Request body"
//	@Success		201		{object}	APIResponse{data=photoOrderResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/photo-order [post]
func (h *PhotoOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid bathhouse ID")
		return
	}

	var req createPhotoOrderRequest
	if err := readJSON(w, r, &req); err != nil {
		return
	}

	region := req.Region
	if region == "" {
		region = "RU"
	}

	order := &domain.PhotoOrder{
		BathhouseID: bathhouseID,
		Region:      region,
		Notes:       req.Notes,
	}

	if err := h.svc.Create(r.Context(), userID, order); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPhotoOrderResponse(order))
}

// ListByOwner godoc
//
//	@Summary		List my photo orders
//	@Tags			photo-orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query	int	false	"Page number"	default(1)
//	@Param			page_size	query	int	false	"Page size"		default(10)
//	@Success		200		{object}	APIResponse{data=[]photoOrderResponse}
//	@Router			/my/photo-orders [get]
func (h *PhotoOrderHandler) ListByOwner(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 10)

	result, err := h.svc.ListByOwner(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	data := make([]photoOrderResponse, len(result.Items))
	for i, o := range result.Items {
		data[i] = toPhotoOrderResponse(&o)
	}

	writeJSONWithMeta(w, http.StatusOK, data, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// GetByID godoc
//
//	@Summary		Get photo order by ID
//	@Tags			photo-orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Order ID"
//	@Success		200	{object}	APIResponse{data=photoOrderResponse}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/photo-orders/{id} [get]
func (h *PhotoOrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid order ID")
		return
	}

	order, err := h.svc.GetByID(r.Context(), userID, role, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPhotoOrderResponse(order))
}

// OwnerCancel godoc
//
//	@Summary		Cancel photo order (owner)
//	@Tags			photo-orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Order ID"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Router			/my/photo-orders/{id}/cancel [post]
func (h *PhotoOrderHandler) OwnerCancel(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid order ID")
		return
	}

	if err := h.svc.OwnerCancel(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "photo order cancelled"})
}

// AdminList godoc
//
//	@Summary		List all photo orders (admin)
//	@Tags			admin-photo-orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status		query	string	false	"Filter by status"
//	@Param			region		query	string	false	"Filter by region"
//	@Param			page		query	int		false	"Page number"	default(1)
//	@Param			page_size	query	int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]photoOrderResponse}
//	@Router			/admin/photo-orders [get]
func (h *PhotoOrderHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	filter := domain.PhotoOrderFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.PhotoOrderStatus(s)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_status", "invalid status filter")
			return
		}
		filter.Status = &status
	}
	if region := r.URL.Query().Get("region"); region != "" {
		filter.Region = &region
	}

	result, err := h.svc.AdminList(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	data := make([]photoOrderResponse, len(result.Items))
	for i, o := range result.Items {
		data[i] = toPhotoOrderResponse(&o)
	}

	writeJSONWithMeta(w, http.StatusOK, data, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

type adminUpdatePhotoOrderRequest struct {
	Status           *string `json:"status,omitempty"`
	PhotographerName *string `json:"photographer_name,omitempty"`
	Price            *int64  `json:"price,omitempty"`
	ScheduledAt      *string `json:"scheduled_at,omitempty"`
	AdminNotes       *string `json:"admin_notes,omitempty"`
}

// AdminUpdate godoc
//
//	@Summary		Update photo order (admin)
//	@Tags			admin-photo-orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string								true	"Order ID"
//	@Param			body	body		adminUpdatePhotoOrderRequest		true	"Update body"
//	@Success		200		{object}	APIResponse{data=photoOrderResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/photo-orders/{id} [put]
func (h *PhotoOrderHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid order ID")
		return
	}

	var req adminUpdatePhotoOrderRequest
	if err := readJSON(w, r, &req); err != nil {
		return
	}

	update := &service.PhotoOrderUpdate{
		PhotographerName: req.PhotographerName,
		Price:            req.Price,
		ScheduledAt:      req.ScheduledAt,
		AdminNotes:       req.AdminNotes,
	}

	if req.Status != nil {
		status := domain.PhotoOrderStatus(*req.Status)
		update.Status = &status
	}

	order, err := h.svc.AdminUpdate(r.Context(), id, update)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPhotoOrderResponse(order))
}

