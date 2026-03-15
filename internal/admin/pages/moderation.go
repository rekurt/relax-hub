package pages

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/logger"
)

var moderationFuncMap = template.FuncMap{
	"stars": func(n int) string {
		if n < 0 {
			n = 0
		} else if n > 5 {
			n = 5
		}
		return strings.Repeat("★", n) + strings.Repeat("☆", 5-n)
	},
	"join":     strings.Join,
	"add":      func(a, b int) int { return a + b },
	"subtract": func(a, b int) int { return a - b },
}

var moderationTmpl = ParsePageTemplate(moderationFuncMap, "templates/moderation.tmpl")

// ModerationReview represents a review in the moderation queue.
type ModerationReview struct {
	ID               string
	UserName         string
	BathhouseID      string
	BathhouseName    string
	Rating           int
	Text             string
	Status           string
	RejectionReasons []string
	Images           []string
	CreatedAt        time.Time
}

// ModerationStats holds moderation statistics.
type ModerationStats struct {
	PendingTotal int64
	ApprovedToday int64
	RejectedToday int64
	PendingWeek   int64
}

// ModerationFilter holds the current filter state for rendering.
type ModerationFilter struct {
	BathhouseID string
	MinRating   string
	MaxRating   string
	FromDate    string
	ToDate      string
	Status      string
}

// BathhouseOption represents a bathhouse for the filter dropdown.
type BathhouseOption struct {
	ID   string
	Name string
}

// ModerationData is the full data model for the moderation page.
type ModerationData struct {
	Reviews     []ModerationReview
	Stats       ModerationStats
	Filter      ModerationFilter
	Bathhouses  []BathhouseOption
	TotalCount  int64
	Page        int
	PageSize    int
	TotalPages  int
	PagesPrefix string
	AdminPrefix string
	PageTitle   string
	ActivePage  string
	GeneratedAt time.Time
}

// ModerationDataProvider fetches moderation data from a data source.
type ModerationDataProvider interface {
	GetModerationData(ctx context.Context, filter ModerationFilter, page, pageSize int) (*ModerationData, error)
	ApproveReview(ctx context.Context, id uuid.UUID) error
	RejectReview(ctx context.Context, id uuid.UUID, reasons []string) error
	BatchApproveReviews(ctx context.Context, ids []uuid.UUID) (successful, failed int)
	BatchRejectReviews(ctx context.Context, ids []uuid.UUID, reasons []string) (successful, failed int)
}

// PostgresModerationProvider fetches moderation data from PostgreSQL.
type PostgresModerationProvider struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewPostgresModerationProvider creates a new PostgresModerationProvider.
func NewPostgresModerationProvider(pool *pgxpool.Pool, log *logger.Logger) *PostgresModerationProvider {
	return &PostgresModerationProvider{pool: pool, log: log}
}

func (p *PostgresModerationProvider) GetModerationData(ctx context.Context, filter ModerationFilter, page, pageSize int) (*ModerationData, error) {
	data := &ModerationData{
		Filter:   filter,
		Page:     page,
		PageSize: pageSize,
	}

	if err := p.loadStats(ctx, &data.Stats); err != nil {
		return nil, err
	}
	if err := p.loadBathhouses(ctx, &data.Bathhouses); err != nil {
		return nil, err
	}
	if err := p.loadReviews(ctx, data); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *PostgresModerationProvider) loadStats(ctx context.Context, stats *ModerationStats) error {
	err := p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE status = 'pending'").Scan(&stats.PendingTotal)
	if err != nil {
		p.log.Error("moderation: count pending", "error", err)
		return err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE status = 'approved' AND updated_at >= $1", todayStart).Scan(&stats.ApprovedToday)
	if err != nil {
		p.log.Error("moderation: count approved today", "error", err)
		return err
	}

	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE status = 'rejected' AND updated_at >= $1", todayStart).Scan(&stats.RejectedToday)
	if err != nil {
		p.log.Error("moderation: count rejected today", "error", err)
		return err
	}

	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	if now.Weekday() == time.Sunday {
		weekStart = todayStart.AddDate(0, 0, -6)
	}
	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE status = 'pending' AND created_at >= $1", weekStart).Scan(&stats.PendingWeek)
	if err != nil {
		p.log.Error("moderation: count pending week", "error", err)
		return err
	}

	return nil
}

func (p *PostgresModerationProvider) loadBathhouses(ctx context.Context, bathhouses *[]BathhouseOption) error {
	rows, err := p.pool.Query(ctx, "SELECT id, name FROM bathhouses ORDER BY name")
	if err != nil {
		p.log.Error("moderation: load bathhouses", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var b BathhouseOption
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
			return err
		}
		*bathhouses = append(*bathhouses, b)
	}
	return rows.Err()
}

func (p *PostgresModerationProvider) loadReviews(ctx context.Context, data *ModerationData) error {
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	status := data.Filter.Status
	if status == "" {
		status = "pending"
	}
	if status != "all" {
		where = append(where, "r.status = $"+strconv.Itoa(argIdx))
		args = append(args, status)
		argIdx++
	}

	if data.Filter.BathhouseID != "" {
		if _, err := uuid.Parse(data.Filter.BathhouseID); err != nil {
			return fmt.Errorf("invalid bathhouse_id filter: %w", err)
		}
		where = append(where, "r.bathhouse_id = $"+strconv.Itoa(argIdx))
		args = append(args, data.Filter.BathhouseID)
		argIdx++
	}

	if data.Filter.MinRating != "" {
		minRating, err := strconv.Atoi(data.Filter.MinRating)
		if err != nil || minRating < 1 || minRating > 5 {
			return fmt.Errorf("invalid min_rating filter")
		}
		where = append(where, "r.rating >= $"+strconv.Itoa(argIdx))
		args = append(args, minRating)
		argIdx++
	}

	if data.Filter.MaxRating != "" {
		maxRating, err := strconv.Atoi(data.Filter.MaxRating)
		if err != nil || maxRating < 1 || maxRating > 5 {
			return fmt.Errorf("invalid max_rating filter")
		}
		where = append(where, "r.rating <= $"+strconv.Itoa(argIdx))
		args = append(args, maxRating)
		argIdx++
	}

	if data.Filter.FromDate != "" {
		if _, err := time.Parse("2006-01-02", data.Filter.FromDate); err != nil {
			return fmt.Errorf("invalid from_date filter")
		}
		where = append(where, "r.created_at >= $"+strconv.Itoa(argIdx))
		args = append(args, data.Filter.FromDate)
		argIdx++
	}

	if data.Filter.ToDate != "" {
		if _, err := time.Parse("2006-01-02", data.Filter.ToDate); err != nil {
			return fmt.Errorf("invalid to_date filter")
		}
		where = append(where, "r.created_at <= $"+strconv.Itoa(argIdx)+"::date + interval '1 day'")
		args = append(args, data.Filter.ToDate)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := "SELECT COUNT(*) FROM reviews r WHERE " + whereClause
	err := p.pool.QueryRow(ctx, countQuery, args...).Scan(&data.TotalCount)
	if err != nil {
		p.log.Error("moderation: count reviews", "error", err)
		return err
	}

	if data.PageSize <= 0 {
		data.PageSize = 20
	}
	data.TotalPages = int(data.TotalCount+int64(data.PageSize)-1) / data.PageSize
	if data.TotalPages == 0 {
		data.TotalPages = 1
	}

	offset := (data.Page - 1) * data.PageSize
	if offset < 0 {
		offset = 0
	}

	// Fetch reviews
	query := `
		SELECT r.id, u.name, r.bathhouse_id, bh.name, r.rating, r.text, r.status, r.rejection_reasons, r.images, r.created_at
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		JOIN bathhouses bh ON bh.id = r.bathhouse_id
		WHERE ` + whereClause + `
		ORDER BY r.created_at DESC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	args = append(args, data.PageSize, offset)

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		p.log.Error("moderation: load reviews", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var mr ModerationReview
		if err := rows.Scan(&mr.ID, &mr.UserName, &mr.BathhouseID, &mr.BathhouseName, &mr.Rating, &mr.Text, &mr.Status, &mr.RejectionReasons, &mr.Images, &mr.CreatedAt); err != nil {
			return err
		}
		data.Reviews = append(data.Reviews, mr)
	}
	return rows.Err()
}

func (p *PostgresModerationProvider) ApproveReview(ctx context.Context, id uuid.UUID) error {
	result, err := p.pool.Exec(ctx, "UPDATE reviews SET status = 'approved', updated_at = NOW() WHERE id = $1 AND status = 'pending'", id)
	if err != nil {
		p.log.Error("moderation: approve review", "error", err, "id", id)
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found or not in pending status")
	}
	return nil
}

func (p *PostgresModerationProvider) RejectReview(ctx context.Context, id uuid.UUID, reasons []string) error {
	result, err := p.pool.Exec(ctx, "UPDATE reviews SET status = 'rejected', rejection_reasons = $2, updated_at = NOW() WHERE id = $1 AND status = 'pending'", id, reasons)
	if err != nil {
		p.log.Error("moderation: reject review", "error", err, "id", id)
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found or not in pending status")
	}
	return nil
}

func (p *PostgresModerationProvider) BatchApproveReviews(ctx context.Context, ids []uuid.UUID) (successful, failed int) {
	if len(ids) == 0 {
		return 0, 0
	}
	result, err := p.pool.Exec(ctx, "UPDATE reviews SET status = 'approved', updated_at = NOW() WHERE id = ANY($1) AND status = 'pending'", ids)
	if err != nil {
		p.log.Error("moderation: batch approve", "error", err, "count", len(ids))
		return 0, len(ids)
	}
	successful = int(result.RowsAffected())
	failed = len(ids) - successful
	return
}

func (p *PostgresModerationProvider) BatchRejectReviews(ctx context.Context, ids []uuid.UUID, reasons []string) (successful, failed int) {
	if len(ids) == 0 {
		return 0, 0
	}
	result, err := p.pool.Exec(ctx, "UPDATE reviews SET status = 'rejected', rejection_reasons = $2, updated_at = NOW() WHERE id = ANY($1) AND status = 'pending'", ids, reasons)
	if err != nil {
		p.log.Error("moderation: batch reject", "error", err, "count", len(ids))
		return 0, len(ids)
	}
	successful = int(result.RowsAffected())
	failed = len(ids) - successful
	return
}

// ModerationHandler serves the moderation page and AJAX endpoints.
type ModerationHandler struct {
	provider    ModerationDataProvider
	log         *logger.Logger
	pagesPrefix string
	adminPrefix string
}

// NewModerationHandler creates a new ModerationHandler.
func NewModerationHandler(provider ModerationDataProvider, log *logger.Logger, pagesPrefix, adminPrefix string) *ModerationHandler {
	return &ModerationHandler{provider: provider, log: log, pagesPrefix: pagesPrefix, adminPrefix: adminPrefix}
}

// ServeHTTP renders the moderation page.
func (h *ModerationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	status := q.Get("status")
	if status == "" {
		status = "pending"
	}

	filter := ModerationFilter{
		BathhouseID: q.Get("bathhouse_id"),
		MinRating:   q.Get("min_rating"),
		MaxRating:   q.Get("max_rating"),
		FromDate:    q.Get("from_date"),
		ToDate:      q.Get("to_date"),
		Status:      status,
	}

	page := 1
	if p := q.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	data, err := h.provider.GetModerationData(r.Context(), filter, page, 20)
	if err != nil {
		h.log.Error("moderation: get data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.PagesPrefix = h.pagesPrefix
	data.AdminPrefix = h.adminPrefix
	data.PageTitle = "Модерация отзывов"
	data.ActivePage = "moderation"
	data.GeneratedAt = time.Now()

	var buf bytes.Buffer
	if err := moderationTmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		h.log.Error("moderation: render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w) //nolint:errcheck
}

type approveRejectRequest struct {
	Reasons []string `json:"reasons"`
}

type batchRequest struct {
	IDs     []string `json:"ids"`
	Reasons []string `json:"reasons"`
}

type actionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type batchResponse struct {
	Success    bool `json:"success"`
	Successful int  `json:"successful"`
	Failed     int  `json:"failed"`
}

// HandleApprove handles AJAX approve request.
func (h *ModerationHandler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "invalid review ID"})
		return
	}

	if err := h.provider.ApproveReview(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, actionResponse{Error: "failed to approve"})
		return
	}

	writeJSON(w, http.StatusOK, actionResponse{Success: true})
}

// HandleReject handles AJAX reject request.
func (h *ModerationHandler) HandleReject(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "invalid review ID"})
		return
	}

	var req approveRejectRequest
	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, actionResponse{Error: "invalid request body"})
			return
		}
	}

	if err := h.provider.RejectReview(r.Context(), id, req.Reasons); err != nil {
		writeJSON(w, http.StatusInternalServerError, actionResponse{Error: "failed to reject"})
		return
	}

	writeJSON(w, http.StatusOK, actionResponse{Success: true})
}

// HandleBatchApprove handles AJAX batch approve request.
func (h *ModerationHandler) HandleBatchApprove(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req batchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "invalid request"})
		return
	}

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "invalid review IDs"})
		return
	}

	if len(ids) == 0 {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "no review IDs provided"})
		return
	}

	if len(ids) > 100 {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "batch size exceeds limit of 100"})
		return
	}

	successful, failed := h.provider.BatchApproveReviews(r.Context(), ids)
	writeJSON(w, http.StatusOK, batchResponse{Success: true, Successful: successful, Failed: failed})
}

// HandleBatchReject handles AJAX batch reject request.
func (h *ModerationHandler) HandleBatchReject(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req batchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "invalid request"})
		return
	}

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "invalid review IDs"})
		return
	}

	if len(ids) == 0 {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "no review IDs provided"})
		return
	}

	if len(ids) > 100 {
		writeJSON(w, http.StatusBadRequest, actionResponse{Error: "batch size exceeds limit of 100"})
		return
	}

	successful, failed := h.provider.BatchRejectReviews(r.Context(), ids, req.Reasons)
	writeJSON(w, http.StatusOK, batchResponse{Success: true, Successful: successful, Failed: failed})
}

func parseUUIDs(strs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
