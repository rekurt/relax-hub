package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type AdminHandler struct {
	userService      service.UserService
	bathhouseService service.BathhouseService
	cityService      service.CityService
	reviewService    service.ReviewService
	notifService     service.NotificationService
}

func NewAdminHandler(
	userService service.UserService,
	bathhouseService service.BathhouseService,
	cityService service.CityService,
	reviewService service.ReviewService,
	notifService service.NotificationService,
) *AdminHandler {
	return &AdminHandler{
		userService:      userService,
		bathhouseService: bathhouseService,
		cityService:      cityService,
		reviewService:    reviewService,
		notifService:     notifService,
	}
}

// ListUsers godoc
//
//	@Summary		List users
//	@Description	Returns a paginated list of all users. Admin only.
//	@Tags			admin-users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]userResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/admin/users [get]
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.userService.List(r.Context(), page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]userResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toUserResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// BlockUser godoc
//
//	@Summary		Block user
//	@Description	Blocks a user account, preventing them from logging in. Cannot block yourself. Admin only.
//	@Tags			admin-users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/users/{id}/block [patch]
func (h *AdminHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return
	}

	adminID := middleware.GetUserID(r.Context())
	if id == adminID {
		writeError(w, http.StatusBadRequest, "invalid_input", "cannot block yourself")
		return
	}

	if err := h.userService.Block(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user blocked"})
}

// UnblockUser godoc
//
//	@Summary		Unblock user
//	@Description	Unblocks a previously blocked user account. Admin only.
//	@Tags			admin-users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/users/{id}/unblock [patch]
func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return
	}

	if err := h.userService.Unblock(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user unblocked"})
}

// ApproveBathhouse godoc
//
//	@Summary		Approve bathhouse
//	@Description	Approves a bathhouse, changing its status to active. Admin only.
//	@Tags			admin-bathhouses
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/bathhouses/{id}/approve [patch]
func (h *AdminHandler) ApproveBathhouse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	if err := h.bathhouseService.Approve(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse approved"})
}

// RejectBathhouse godoc
//
//	@Summary		Reject bathhouse
//	@Description	Rejects a bathhouse registration. Admin only.
//	@Tags			admin-bathhouses
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/bathhouses/{id}/reject [patch]
func (h *AdminHandler) RejectBathhouse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	if err := h.bathhouseService.Reject(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse rejected"})
}

// ListBathhouses godoc
//
//	@Summary		List bathhouses (admin)
//	@Description	Returns a paginated list of bathhouses with optional status filter. Shows all statuses by default. Admin only.
//	@Tags			admin-bathhouses
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Param			status		query		string	false	"Filter by status (pending, active, rejected, blocked)"
//	@Success		200			{object}	APIResponse{data=[]bathhouseResponse,meta=Meta}
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/admin/bathhouses [get]
func (h *AdminHandler) ListBathhouses(w http.ResponseWriter, r *http.Request) {
	filter := domain.BathhouseFilter{
		Page:     getPage(r.URL.Query().Get("page")),
		PageSize: getPageSize(r.URL.Query().Get("page_size"), 20),
	}

	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.BathhouseStatus(v)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid status value")
			return
		}
		filter.Status = &status
	} else {
		filter.ShowAllStatuses = true
	}

	result, err := h.bathhouseService.Search(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bathhouseResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBathhouseResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

