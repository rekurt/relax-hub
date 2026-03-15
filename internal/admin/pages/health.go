package pages

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/redis/go-redis/v9"
)

var healthTmpl = ParsePageTemplate(nil, "templates/health.tmpl")

// ServiceStatus represents the health of a single external service.
type ServiceStatus struct {
	Name    string
	Status  string // "up" or "down"
	Latency time.Duration
	Error   string
}

// ModerationBacklog represents the aging of pending reviews.
type ModerationBacklog struct {
	Over24h int64
	Over48h int64
	Over72h int64
}

// SubscriptionHealth holds expiring subscription counts.
type SubscriptionHealth struct {
	ExpiringThisWeek int64
}

// HealthData is the full data model for the platform health monitor page.
type HealthData struct {
	Services           []ServiceStatus
	WebSocketClients   int
	ModerationBacklog  ModerationBacklog
	SubscriptionHealth SubscriptionHealth
	GeneratedAt        time.Time
	PagesPrefix        string
	AdminPrefix        string
	PageTitle          string
	ActivePage         string
}

// HealthDataProvider fetches health data from external services.
type HealthDataProvider interface {
	GetHealthData(ctx context.Context) (*HealthData, error)
}

// PlatformHealthProvider checks real service health.
type PlatformHealthProvider struct {
	pool  *pgxpool.Pool
	redis *redis.Client
	hub   *notification.Hub
	log   *logger.Logger
}

// NewPlatformHealthProvider creates a new PlatformHealthProvider.
func NewPlatformHealthProvider(pool *pgxpool.Pool, redisClient *redis.Client, hub *notification.Hub, log *logger.Logger) *PlatformHealthProvider {
	return &PlatformHealthProvider{pool: pool, redis: redisClient, hub: hub, log: log}
}

func (p *PlatformHealthProvider) GetHealthData(ctx context.Context) (*HealthData, error) {
	data := &HealthData{GeneratedAt: time.Now()}

	data.Services = p.checkServices(ctx)

	if p.hub != nil {
		data.WebSocketClients = p.hub.OnlineCount()
	}

	if err := p.loadModerationBacklog(ctx, &data.ModerationBacklog); err != nil {
		p.log.Error("health: moderation backlog", "error", err)
	}

	if err := p.loadSubscriptionHealth(ctx, &data.SubscriptionHealth); err != nil {
		p.log.Error("health: subscription health", "error", err)
	}

	return data, nil
}

func (p *PlatformHealthProvider) checkServices(ctx context.Context) []ServiceStatus {
	var services []ServiceStatus

	// PostgreSQL
	pgStatus := ServiceStatus{Name: "PostgreSQL"}
	start := time.Now()
	pgCtx, pgCancel := context.WithTimeout(ctx, 3*time.Second)
	defer pgCancel()
	if err := p.pool.Ping(pgCtx); err != nil {
		pgStatus.Status = "down"
		pgStatus.Error = err.Error()
	} else {
		pgStatus.Status = "up"
	}
	pgStatus.Latency = time.Since(start)
	services = append(services, pgStatus)

	// Redis
	redisStatus := ServiceStatus{Name: "Redis"}
	start = time.Now()
	if p.redis != nil {
		redisCtx, redisCancel := context.WithTimeout(ctx, 3*time.Second)
		defer redisCancel()
		if err := p.redis.Ping(redisCtx).Err(); err != nil {
			redisStatus.Status = "down"
			redisStatus.Error = err.Error()
		} else {
			redisStatus.Status = "up"
		}
	} else {
		redisStatus.Status = "unconfigured"
	}
	redisStatus.Latency = time.Since(start)
	services = append(services, redisStatus)

	return services
}

func (p *PlatformHealthProvider) loadModerationBacklog(ctx context.Context, backlog *ModerationBacklog) error {
	now := time.Now()
	h24 := now.Add(-24 * time.Hour)
	h48 := now.Add(-48 * time.Hour)
	h72 := now.Add(-72 * time.Hour)

	err := p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE status = 'pending' AND created_at <= $1", h24).Scan(&backlog.Over24h)
	if err != nil {
		return err
	}
	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE status = 'pending' AND created_at <= $1", h48).Scan(&backlog.Over48h)
	if err != nil {
		return err
	}
	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE status = 'pending' AND created_at <= $1", h72).Scan(&backlog.Over72h)
	return err
}

func (p *PlatformHealthProvider) loadSubscriptionHealth(ctx context.Context, health *SubscriptionHealth) error {
	now := time.Now()
	weekEnd := now.AddDate(0, 0, 7)
	return p.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM subscriptions WHERE status = 'active' AND end_date <= $1 AND end_date >= $2",
		weekEnd, now,
	).Scan(&health.ExpiringThisWeek)
}

// HealthHandler serves the platform health monitor page.
type HealthHandler struct {
	provider    HealthDataProvider
	log         *logger.Logger
	pagesPrefix string
	adminPrefix string
}

// NewHealthHandler creates a new admin HealthHandler.
func NewHealthHandler(provider HealthDataProvider, log *logger.Logger, pagesPrefix, adminPrefix string) *HealthHandler {
	return &HealthHandler{provider: provider, log: log, pagesPrefix: pagesPrefix, adminPrefix: adminPrefix}
}

// ServeHTTP renders the platform health monitor page.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h.provider.GetHealthData(r.Context())
	if err != nil {
		h.log.Error("health: get data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.PagesPrefix = h.pagesPrefix
	data.AdminPrefix = h.adminPrefix
	data.PageTitle = "Мониторинг платформы"
	data.ActivePage = "health"

	var buf bytes.Buffer
	if err := healthTmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		h.log.Error("health: render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w) //nolint:errcheck
}
