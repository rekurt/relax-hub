package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
	log              *logger.Logger
}

func NewAnalyticsHandler(
	analyticsService service.AnalyticsService,
	log *logger.Logger,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		log:              log,
	}
}

// GetOwnerDashboard returns analytics dashboard for a bathhouse owner
// GET /api/v1/my/bathhouses/{id}/analytics?period=30d
func (h *AnalyticsHandler) GetOwnerDashboard(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		periodStr = "30d" // default to 30 days
	}

	period := domain.AnalyticsPeriod(periodStr)
	if !period.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid period, must be one of: 1d, 7d, 30d, 90d")
		return
	}

	dashboard, err := h.analyticsService.GetOwnerDashboard(r.Context(), userID, userRole, id, period)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dashboard)
}

// GetOwnerDailyStats returns daily analytics breakdown for a bathhouse
// GET /api/v1/my/bathhouses/{id}/analytics/daily?from=2024-01-01&to=2024-01-31
func (h *AnalyticsHandler) GetOwnerDailyStats(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	// Parse date parameters
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "from and to date parameters required")
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid from date format (use YYYY-MM-DD)")
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid to date format (use YYYY-MM-DD)")
		return
	}

	// Get daily stats from service (RBAC check is done in service layer)
	dailyStats, err := h.analyticsService.GetDailyStats(r.Context(), userID, userRole, id, from, to)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": dailyStats,
	})
}

// GetAdminDashboard returns platform-wide analytics dashboard
// GET /api/v1/admin/analytics?period=30d
func (h *AnalyticsHandler) GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	userRole := middleware.GetUserRole(r.Context())

	// RBAC check: only admins can view platform analytics
	if userRole != domain.RoleAdmin {
		writeError(w, http.StatusForbidden, "forbidden", "admin role required")
		return
	}

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		periodStr = "30d" // default to 30 days
	}

	period := domain.AnalyticsPeriod(periodStr)
	if !period.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid period, must be one of: 1d, 7d, 30d, 90d")
		return
	}

	dashboard, err := h.analyticsService.GetAdminDashboard(r.Context(), userRole, period)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dashboard)
}

// GetTopBathhouses returns top bathhouses ranked by a metric
// GET /api/v1/admin/analytics/top?metric=bookings&limit=10
func (h *AnalyticsHandler) GetTopBathhouses(w http.ResponseWriter, r *http.Request) {
	userRole := middleware.GetUserRole(r.Context())

	// RBAC check: only admins can view platform analytics
	if userRole != domain.RoleAdmin {
		writeError(w, http.StatusForbidden, "forbidden", "admin role required")
		return
	}

	metricStr := r.URL.Query().Get("metric")
	if metricStr == "" {
		metricStr = "bookings" // default metric
	}

	metric := domain.TopMetric(metricStr)
	if !metric.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid metric, must be one of: views, bookings, revenue, rating")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	// Get top bathhouses for the specified metric
	topBathhouses, err := h.analyticsService.GetTopBathhousesByMetric(r.Context(), userRole, metric, int64(limit))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metric":      metricStr,
		"limit":       limit,
		"bathhouses":  topBathhouses,
		"total_count": len(topBathhouses),
	})
}
