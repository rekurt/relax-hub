package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/service"
)

type AdminRoleHandler struct {
	svc service.AdminRoleService
}

func NewAdminRoleHandler(svc service.AdminRoleService) *AdminRoleHandler {
	return &AdminRoleHandler{svc: svc}
}

type setAdminSubRoleRequest struct {
	SubRole domain.AdminSubRole `json:"sub_role"`
}

type adminRoleResponse struct {
	ID           uuid.UUID           `json:"id"`
	Email        string              `json:"email"`
	Name         string              `json:"name"`
	AdminSubRole domain.AdminSubRole `json:"admin_sub_role"`
	TwoFAMethod  domain.TwoFAMethod  `json:"two_fa_method"`
	IsActive     bool                `json:"is_active"`
}

type permissionsMatrixResponse struct {
	Roles       []domain.AdminSubRole                            `json:"roles"`
	Permissions []domain.AdminPermission                         `json:"permissions"`
	Matrix      map[domain.AdminSubRole][]domain.AdminPermission `json:"matrix"`
}

// ListAdminUsers godoc
//
//	@Summary		List admin users with sub-roles
//	@Description	Returns all users with admin role and their sub-roles. Super admin only.
//	@Tags			admin-roles
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]adminRoleResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/admin/roles [get]
func (h *AdminRoleHandler) ListAdminUsers(w http.ResponseWriter, r *http.Request) {
	admins, err := h.svc.ListAdminUsers(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := make([]adminRoleResponse, len(admins))
	for i, a := range admins {
		resp[i] = adminRoleResponse{
			ID:           a.ID,
			Email:        a.Email,
			Name:         a.Name,
			AdminSubRole: a.AdminSubRole,
			TwoFAMethod:  a.TwoFAMethod,
			IsActive:     a.IsActive,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// SetAdminSubRole godoc
//
//	@Summary		Set admin sub-role
//	@Description	Assigns a sub-role to an admin user. Super admin only.
//	@Tags			admin-roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"User ID"
//	@Param			body	body		setAdminSubRoleRequest	true	"Sub-role to assign"
//	@Success		200		{object}	APIResponse{data=statusResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/roles/{id} [put]
func (h *AdminRoleHandler) SetAdminSubRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	var req setAdminSubRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if !req.SubRole.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_sub_role", "invalid admin sub-role")
		return
	}

	if err := h.svc.SetAdminSubRole(r.Context(), userID, req.SubRole); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetPermissionsMatrix godoc
//
//	@Summary		Get admin permissions matrix
//	@Description	Returns the full permissions matrix showing which sub-roles have which permissions.
//	@Tags			admin-roles
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=permissionsMatrixResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/admin/roles/permissions [get]
func (h *AdminRoleHandler) GetPermissionsMatrix(w http.ResponseWriter, _ *http.Request) {
	matrix := h.svc.GetPermissionsMatrix()
	resp := permissionsMatrixResponse{
		Roles:       domain.AllAdminSubRoles(),
		Permissions: domain.AllAdminPermissions(),
		Matrix:      matrix,
	}
	writeJSON(w, http.StatusOK, resp)
}
