package pages

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/redis/go-redis/v9"
)

var healthFuncMap = template.FuncMap{
	"formatLatency":   FormatLatency,
	"backlogSeverity": BacklogSeverity,
	"formatBytes":     FormatBytes,
	"formatUptime":    FormatUptime,
	"poolUsageClass":  PoolUsageClass,
	"poolPercent":     PoolPercent,
}

var healthTmpl = ParsePageTemplate(healthFuncMap, "templates/health.tmpl")

// FormatLatency formats a time.Duration to a human-readable string like "2 ms" or "150 ms".
func FormatLatency(d time.Duration) string {
	us := d.Microseconds()
	if us < 1000 {
		return fmt.Sprintf("%d µs", us)
	}
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms)
	}
	return fmt.Sprintf("%.1f s", d.Seconds())
}

// BacklogSeverity returns a CSS class based on backlog count thresholds.
func BacklogSeverity(count int64) string {
	switch {
	case count == 0:
		return "severity-ok"
	case count <= 5:
		return "severity-warning"
	case count <= 20:
		return "severity-elevated"
	default:
		return "severity-critical"
	}
}

// FormatBytes formats byte count to human-readable string (KB, MB, GB).
func FormatBytes(bytes uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatUptime formats a duration to a human-readable Russian uptime string.
func FormatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d дн. %d ч. %d мин.", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d ч. %d мин.", hours, minutes)
	}
	return fmt.Sprintf("%d мин.", minutes)
}

// PoolUsageClass returns a CSS class based on connection pool usage percentage.
func PoolUsageClass(current, max int32) string {
	if max == 0 {
		return "pool-ok"
	}
	pct := float64(current) / float64(max) * 100
	switch {
	case pct < 50:
		return "pool-ok"
	case pct < 80:
		return "pool-warning"
	default:
		return "pool-danger"
	}
}

// PoolPercent calculates the percentage of pool usage.
func PoolPercent(current, max int32) int {
	if max == 0 {
		return 0
	}
	pct := int(float64(current) / float64(max) * 100)
	if pct > 100 {
		pct = 100
	}
	return pct
}

// ServiceStatus represents the health of a single external service.
type ServiceStatus struct {
	Name    string
	Status  string // "up", "down", or "unconfigured"
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

// SystemMetrics holds runtime system metrics.
type SystemMetrics struct {
	Uptime        time.Duration
	GoVersion     string
	NumGoroutines int
	NumCPU        int
	MemAlloc      uint64 // bytes currently allocated
	MemTotalAlloc uint64 // cumulative bytes allocated
	MemSys        uint64 // bytes obtained from system
	NumGC         uint32 // number of completed GC cycles
}

// PoolStats holds pgxpool connection statistics.
type PoolStats struct {
	AcquiredConns     int32
	IdleConns         int32
	TotalConns        int32
	MaxConns          int32
	AcquireCount      int64
	EmptyAcquireCount int64
	AcquireDuration   time.Duration
}

// HealthData is the full data model for the platform health monitor page.
type HealthData struct {
	Services           []ServiceStatus
	WebSocketClients   int
	ModerationBacklog  ModerationBacklog
	SubscriptionHealth SubscriptionHealth
	System             SystemMetrics
	Pool               PoolStats
	DatabaseSize       string
	FilesystemOK       bool
	FilesystemError    string
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
	pool      *pgxpool.Pool
	redis     *redis.Client
	hub       *notification.Hub
	log       *logger.Logger
	startTime time.Time
}

// NewPlatformHealthProvider creates a new PlatformHealthProvider.
func NewPlatformHealthProvider(pool *pgxpool.Pool, redisClient *redis.Client, hub *notification.Hub, log *logger.Logger) *PlatformHealthProvider {
	return &PlatformHealthProvider{pool: pool, redis: redisClient, hub: hub, log: log, startTime: time.Now()}
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

	data.System = p.collectSystemMetrics()
	data.Pool = p.collectPoolStats()

	if err := p.loadDatabaseSize(ctx, data); err != nil {
		p.log.Error("health: database size", "error", err)
	}

	p.checkFilesystem(data)

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

func (p *PlatformHealthProvider) collectSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return SystemMetrics{
		Uptime:        time.Since(p.startTime),
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		MemAlloc:      m.Alloc,
		MemTotalAlloc: m.TotalAlloc,
		MemSys:        m.Sys,
		NumGC:         m.NumGC,
	}
}

func (p *PlatformHealthProvider) collectPoolStats() PoolStats {
	stat := p.pool.Stat()
	return PoolStats{
		AcquiredConns:     stat.AcquiredConns(),
		IdleConns:         stat.IdleConns(),
		TotalConns:        stat.TotalConns(),
		MaxConns:          stat.MaxConns(),
		AcquireCount:      stat.AcquireCount(),
		EmptyAcquireCount: stat.EmptyAcquireCount(),
		AcquireDuration:   stat.AcquireDuration(),
	}
}

func (p *PlatformHealthProvider) loadDatabaseSize(ctx context.Context, data *HealthData) error {
	var size string
	err := p.pool.QueryRow(ctx,
		"SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&size)
	if err != nil {
		return err
	}
	data.DatabaseSize = size
	return nil
}

func (p *PlatformHealthProvider) checkFilesystem(data *HealthData) {
	f, err := os.CreateTemp("", "bani-health-check-*.tmp")
	if err != nil {
		data.FilesystemOK = false
		data.FilesystemError = err.Error()
		return
	}
	name := f.Name()
	f.Close()
	os.Remove(name) //nolint:errcheck
	data.FilesystemOK = true
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
