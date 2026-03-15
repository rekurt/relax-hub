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

// GetOwnerDashboard godoc
// @Summary      Get owner analytics dashboard
// @Description  Returns analytics dashboard for a bathhouse owner, including booking stats, revenue, and views. Owner or representative only.
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true   "Bathhouse ID (UUID)"
// @Param        period  query     string  false  "Period: 1d, 7d, 30d, 90d"  default(30d)
// @Success      200     {object}  APIResponse{data=object}
// @Failure      400     {object}  APIResponse{error=APIError}
// @Failure      401     {object}  APIResponse{error=APIError}
// @Failure      403     {object}  APIResponse{error=APIError}
// @Failure      404     {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/analytics [get]
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

// GetOwnerDailyStats godoc
// @Summary      Get daily analytics
// @Description  Returns daily analytics breakdown for a bathhouse within a date range (max 365 days). Owner or representative only.
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Bathhouse ID (UUID)"
// @Param        from  query     string  true  "Start date (YYYY-MM-DD)"
// @Param        to    query     string  true  "End date (YYYY-MM-DD)"
// @Success      200   {object}  APIResponse{data=object}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/analytics/daily [get]
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

	// Validate date range: from must be before to
	if from.After(to) {
		writeError(w, http.StatusBadRequest, "invalid_input", "from date must be before to date")
		return
	}

	// Prevent excessive date ranges (max 365 days) to prevent DoS
	maxDays := 365 * time.Hour * 24
	if to.Sub(from) > maxDays {
		writeError(w, http.StatusBadRequest, "invalid_input", "date range cannot exceed 365 days")
		return
	}

	// Get daily stats from service (RBAC check is done in service layer)
	dailyStats, err := h.analyticsService.GetDailyStats(r.Context(), userID, userRole, id, from, to)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dailyStats)
}

// GetAdminDashboard godoc
// @Summary      Get admin analytics dashboard
// @Description  Returns platform-wide analytics dashboard. Admin only.
// @Tags         admin-analytics
// @Produce      json
// @Security     BearerAuth
// @Param        period  query     string  false  "Period: 1d, 7d, 30d, 90d"  default(30d)
// @Success      200     {object}  APIResponse{data=object}
// @Failure      400     {object}  APIResponse{error=APIError}
// @Failure      401     {object}  APIResponse{error=APIError}
// @Failure      403     {object}  APIResponse{error=APIError}
// @Router       /admin/analytics [get]
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

// GetTopBathhouses godoc
// @Summary      Get top bathhouses
// @Description  Returns top bathhouses ranked by a specified metric (views, bookings, revenue, rating). Admin only.
// @Tags         admin-analytics
// @Produce      json
// @Security     BearerAuth
// @Param        metric  query     string  false  "Metric: views, bookings, revenue, rating"  default(bookings)
// @Param        limit   query     int     false  "Max results (1-100)"                        default(10)
// @Success      200     {object}  APIResponse{data=object}
// @Failure      400     {object}  APIResponse{error=APIError}
// @Failure      401     {object}  APIResponse{error=APIError}
// @Failure      403     {object}  APIResponse{error=APIError}
// @Router       /admin/analytics/top [get]
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
