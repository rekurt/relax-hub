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

type PMSHandler struct {
	pmsService service.PMSService
}

func NewPMSHandler(pmsService service.PMSService) *PMSHandler {
	return &PMSHandler{pmsService: pmsService}
}

type createPMSConnectionRequest struct {
	BathhouseID         string `json:"bathhouse_id"`
	Provider            string `json:"provider"`
	Credentials         string `json:"credentials"`
	SyncDirection       string `json:"sync_direction"`
	SyncIntervalMinutes int    `json:"sync_interval_minutes"`
	ExternalID          string `json:"external_id"`
}

type updatePMSConnectionRequest struct {
	Provider            string `json:"provider"`
	Credentials         string `json:"credentials"`
	SyncDirection       string `json:"sync_direction"`
	SyncIntervalMinutes int    `json:"sync_interval_minutes"`
	ExternalID          string `json:"external_id"`
	Status              string `json:"status"`
}

type pmsConnectionResponse struct {
	ID                  string  `json:"id"`
	OwnerID             string  `json:"owner_id"`
	BathhouseID         string  `json:"bathhouse_id"`
	Provider            string  `json:"provider"`
	SyncDirection       string  `json:"sync_direction"`
	SyncIntervalMinutes int     `json:"sync_interval_minutes"`
	Status              string  `json:"status"`
	LastSyncAt          *string `json:"last_sync_at,omitempty"`
	LastSyncError       string  `json:"last_sync_error,omitempty"`
	ExternalID          string  `json:"external_id"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

type pmsSyncLogResponse struct {
	ID           string `json:"id"`
	ConnectionID string `json:"connection_id"`
	Direction    string `json:"direction"`
	Status       string `json:"status"`
	ItemsSynced  int    `json:"items_synced"`
	ErrorMessage string `json:"error_message,omitempty"`
	StartedAt    string `json:"started_at"`
	CompletedAt  string `json:"completed_at"`
}

func toPMSConnectionResponse(c *domain.PMSConnection) pmsConnectionResponse {
	resp := pmsConnectionResponse{
		ID:                  c.ID.String(),
		OwnerID:             c.OwnerID.String(),
		BathhouseID:         c.BathhouseID.String(),
		Provider:            string(c.Provider),
		SyncDirection:       string(c.SyncDirection),
		SyncIntervalMinutes: c.SyncIntervalMinutes,
		Status:              string(c.Status),
		LastSyncError:       c.LastSyncError,
		ExternalID:          c.ExternalID,
		CreatedAt:           c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           c.UpdatedAt.Format(time.RFC3339),
	}
	if c.LastSyncAt != nil {
		s := c.LastSyncAt.Format(time.RFC3339)
		resp.LastSyncAt = &s
	}
	return resp
}

func toPMSSyncLogResponse(l *domain.PMSSyncLog) pmsSyncLogResponse {
	return pmsSyncLogResponse{
		ID:           l.ID.String(),
		ConnectionID: l.ConnectionID.String(),
		Direction:    string(l.Direction),
		Status:       l.Status,
		ItemsSynced:  l.ItemsSynced,
		ErrorMessage: l.ErrorMessage,
		StartedAt:    l.StartedAt.Format(time.RFC3339),
		CompletedAt:  l.CompletedAt.Format(time.RFC3339),
	}
}

// CreatePMSConnection godoc
//
//	@Summary		Create a PMS connection
//	@Description	Connect a bathhouse to an external PMS (Yclients or Restoplace)
//	@Tags			pms
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createPMSConnectionRequest	true	"PMS connection data"
//	@Success		201		{object}	APIResponse{data=pmsConnectionResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/pms-connections [post]
func (h *PMSHandler) CreatePMSConnection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var req createPMSConnectionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	bathhouseID, err := uuid.Parse(req.BathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_bathhouse_id", "invalid bathhouse ID")
		return
	}

	syncInterval := req.SyncIntervalMinutes
	if syncInterval == 0 {
		syncInterval = domain.DefaultPMSSyncIntervalMin
	}

	syncDirection := domain.PMSSyncDirection(req.SyncDirection)
	if syncDirection == "" {
		syncDirection = domain.PMSSyncDirectionBoth
	}

	conn := &domain.PMSConnection{
		ID:                   uuid.New(),
		BathhouseID:          bathhouseID,
		Provider:             domain.PMSProvider(req.Provider),
		CredentialsEncrypted: req.Credentials,
		SyncDirection:        syncDirection,
		SyncIntervalMinutes:  syncInterval,
		ExternalID:           req.ExternalID,
	}

	if err := h.pmsService.Create(r.Context(), userID, role, conn); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPMSConnectionResponse(conn))
}

// ListPMSConnections godoc
//
//	@Summary		List PMS connections
//	@Description	List all PMS connections for the authenticated owner
//	@Tags			pms
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]pmsConnectionResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/my/pms-connections [get]
func (h *PMSHandler) ListPMSConnections(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.pmsService.List(r.Context(), userID, role, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []pmsConnectionResponse
	for _, c := range result.Items {
		items = append(items, toPMSConnectionResponse(&c))
	}
	if items == nil {
		items = []pmsConnectionResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: int((result.TotalCount + int64(result.PageSize) - 1) / int64(result.PageSize)),
	})
}

// GetPMSConnection godoc
//
//	@Summary		Get a PMS connection
//	@Description	Get PMS connection details by ID
//	@Tags			pms
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Connection ID (UUID)"
//	@Success		200	{object}	APIResponse{data=pmsConnectionResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/pms-connections/{id} [get]
func (h *PMSHandler) GetPMSConnection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid connection ID")
		return
	}

	conn, err := h.pmsService.GetByID(r.Context(), userID, role, connID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPMSConnectionResponse(conn))
}

// UpdatePMSConnection godoc
//
//	@Summary		Update a PMS connection
//	@Description	Update an existing PMS connection
//	@Tags			pms
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Connection ID (UUID)"
//	@Param			body	body		updatePMSConnectionRequest	true	"Updated connection data"
//	@Success		200		{object}	APIResponse{data=pmsConnectionResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/pms-connections/{id} [put]
func (h *PMSHandler) UpdatePMSConnection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid connection ID")
		return
	}

	var req updatePMSConnectionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	syncInterval := req.SyncIntervalMinutes
	if syncInterval == 0 {
		syncInterval = domain.DefaultPMSSyncIntervalMin
	}

	syncDirection := domain.PMSSyncDirection(req.SyncDirection)
	if syncDirection == "" {
		syncDirection = domain.PMSSyncDirectionBoth
	}

	status := domain.PMSConnectionStatus(req.Status)
	if status == "" {
		status = domain.PMSConnectionActive
	}

	conn := &domain.PMSConnection{
		ID:                   connID,
		Provider:             domain.PMSProvider(req.Provider),
		CredentialsEncrypted: req.Credentials,
		SyncDirection:        syncDirection,
		SyncIntervalMinutes:  syncInterval,
		ExternalID:           req.ExternalID,
		Status:               status,
	}

	if err := h.pmsService.Update(r.Context(), userID, role, conn); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPMSConnectionResponse(conn))
}

// DeletePMSConnection godoc
//
//	@Summary		Delete a PMS connection
//	@Description	Delete a PMS connection
//	@Tags			pms
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Connection ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/pms-connections/{id} [delete]
func (h *PMSHandler) DeletePMSConnection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid connection ID")
		return
	}

	if err := h.pmsService.Delete(r.Context(), userID, role, connID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "PMS connection deleted"})
}

// TestPMSConnection godoc
//
//	@Summary		Test a PMS connection
//	@Description	Test connectivity to the external PMS
//	@Tags			pms
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Connection ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/pms-connections/{id}/test [post]
func (h *PMSHandler) TestPMSConnection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid connection ID")
		return
	}

	if err := h.pmsService.TestConnection(r.Context(), userID, role, connID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "PMS connection test successful"})
}

// SyncPMSConnection godoc
//
//	@Summary		Trigger PMS sync
//	@Description	Manually trigger synchronization with the external PMS
//	@Tags			pms
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Connection ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/pms-connections/{id}/sync [post]
func (h *PMSHandler) SyncPMSConnection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid connection ID")
		return
	}

	if err := h.pmsService.TriggerSync(r.Context(), userID, role, connID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "PMS sync completed"})
}

// ListPMSSyncLogs godoc
//
//	@Summary		List PMS sync logs
//	@Description	List synchronization logs for a PMS connection
//	@Tags			pms
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Connection ID (UUID)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]pmsSyncLogResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Failure		404			{object}	APIResponse{error=APIError}
//	@Router			/my/pms-connections/{id}/logs [get]
func (h *PMSHandler) ListPMSSyncLogs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid connection ID")
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

	result, err := h.pmsService.ListSyncLogs(r.Context(), userID, role, connID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []pmsSyncLogResponse
	for _, l := range result.Items {
		items = append(items, toPMSSyncLogResponse(&l))
	}
	if items == nil {
		items = []pmsSyncLogResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: int((result.TotalCount + int64(result.PageSize) - 1) / int64(result.PageSize)),
	})
}
