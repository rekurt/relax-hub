package pages

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
)

// FunnelStepRu translates funnel step names to Russian.
func FunnelStepRu(name string) string {
	switch name {
	case "visit":
		return "Визит"
	case "search":
		return "Поиск"
	case "view_card":
		return "Просмотр карточки"
	case "start_booking":
		return "Начало бронирования"
	case "pay":
		return "Оплата"
	case "complete_visit":
		return "Завершённый визит"
	default:
		return name
	}
}

// CohortCellClass returns a CSS class for a cohort retention cell based on the percentage.
func CohortCellClass(pct float64) string {
	switch {
	case pct >= 50:
		return "cohort-cell-high"
	case pct >= 20:
		return "cohort-cell-mid"
	case pct > 0:
		return "cohort-cell-low"
	default:
		return "cohort-cell-none"
	}
}

var advancedAnalyticsFuncMap = template.FuncMap{
	"formatRubles": FormatKopecksToRubles,
	"formatPct": func(f float64) string {
		return strconv.FormatFloat(f, 'f', 1, 64) + "%"
	},
	"funnelStepRu":   FunnelStepRu,
	"cohortCellClass": CohortCellClass,
	"divFloat": func(a, b int64) float64 {
		if b == 0 {
			return 0
		}
		return float64(a) / float64(b)
	},
	"list": func(items ...string) []string {
		return items
	},
}

var advancedAnalyticsTmpl = ParsePageTemplate(advancedAnalyticsFuncMap, "templates/advanced_analytics.tmpl")

// AdvancedAnalyticsData is the data model for the advanced analytics page.
type AdvancedAnalyticsData struct {
	Funnel        *domain.ConversionFunnel  `json:"funnel"`
	Cohorts       *domain.CohortAnalysis    `json:"cohorts"`
	GeoDemand     *domain.GeoDemandSupplyMap `json:"geo_demand"`
	WalletMetrics *domain.WalletMetrics     `json:"wallet_metrics"`
	Period        string                    `json:"period"`
	GeneratedAt   time.Time                 `json:"generated_at"`
	PagesPrefix   string
	AdminPrefix   string
	PageTitle     string
	ActivePage    string
}

// AdvancedAnalyticsDataProvider fetches advanced analytics data.
type AdvancedAnalyticsDataProvider interface {
	GetAdvancedAnalyticsData(ctx context.Context, period string) (*AdvancedAnalyticsData, error)
}

// AdvancedAnalyticsQuerier is the subset of AnalyticsRepository used by the admin page.
type AdvancedAnalyticsQuerier interface {
	GetConversionFunnel(ctx context.Context, from, to time.Time) ([]domain.FunnelStep, error)
	GetCohortAnalysis(ctx context.Context, months int) ([]domain.CohortRow, error)
	GetGeoSupplyDemand(ctx context.Context, from, to time.Time) ([]domain.GeoSupplyDemand, error)
	GetWalletMetrics(ctx context.Context, from, to time.Time) (*domain.WalletMetrics, error)
}

// PostgresAdvancedAnalyticsProvider fetches advanced analytics from PostgreSQL via the analytics repository.
type PostgresAdvancedAnalyticsProvider struct {
	repo AdvancedAnalyticsQuerier
	log  *logger.Logger
}

// NewPostgresAdvancedAnalyticsProvider creates a provider that wraps a repository.AnalyticsRepository.
func NewPostgresAdvancedAnalyticsProvider(repo AdvancedAnalyticsQuerier, log *logger.Logger) *PostgresAdvancedAnalyticsProvider {
	return &PostgresAdvancedAnalyticsProvider{repo: repo, log: log}
}

func (p *PostgresAdvancedAnalyticsProvider) GetAdvancedAnalyticsData(ctx context.Context, periodStr string) (*AdvancedAnalyticsData, error) {
	period := domain.AnalyticsPeriod(periodStr)
	if !period.IsValid() {
		period = domain.PeriodMonth
	}

	now := time.Now()
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	from := to.AddDate(0, 0, -period.Days())
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	data := &AdvancedAnalyticsData{
		Period:      periodStr,
		GeneratedAt: now,
	}

	funnel, err := p.repo.GetConversionFunnel(ctx, from, to)
	if err != nil {
		p.log.Warn("advanced analytics: funnel", "error", err)
	} else {
		data.Funnel = &domain.ConversionFunnel{Period: period, Steps: funnel}
	}

	cohorts, err := p.repo.GetCohortAnalysis(ctx, 6)
	if err != nil {
		p.log.Warn("advanced analytics: cohorts", "error", err)
	} else {
		data.Cohorts = &domain.CohortAnalysis{Cohorts: cohorts}
	}

	geo, err := p.repo.GetGeoSupplyDemand(ctx, from, to)
	if err != nil {
		p.log.Warn("advanced analytics: geo", "error", err)
	} else {
		data.GeoDemand = &domain.GeoDemandSupplyMap{Period: period, Cities: geo}
	}

	walletMetrics, err := p.repo.GetWalletMetrics(ctx, from, to)
	if err != nil {
		p.log.Warn("advanced analytics: wallet", "error", err)
	} else {
		data.WalletMetrics = walletMetrics
	}

	return data, nil
}

// AdvancedAnalyticsPageHandler serves the advanced analytics dashboard page.
type AdvancedAnalyticsPageHandler struct {
	provider    AdvancedAnalyticsDataProvider
	log         *logger.Logger
	pagesPrefix string
	adminPrefix string
}

// NewAdvancedAnalyticsHandler creates a new AdvancedAnalyticsPageHandler.
func NewAdvancedAnalyticsHandler(provider AdvancedAnalyticsDataProvider, log *logger.Logger, pagesPrefix, adminPrefix string) *AdvancedAnalyticsPageHandler {
	return &AdvancedAnalyticsPageHandler{provider: provider, log: log, pagesPrefix: pagesPrefix, adminPrefix: adminPrefix}
}

// ServeHTTP renders the advanced analytics dashboard page.
func (h *AdvancedAnalyticsPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	data, err := h.provider.GetAdvancedAnalyticsData(r.Context(), period)
	if err != nil {
		h.log.Error("advanced analytics: get data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.PagesPrefix = h.pagesPrefix
	data.AdminPrefix = h.adminPrefix
	data.PageTitle = "Расширенная аналитика"
	data.ActivePage = "advanced-analytics"

	var buf bytes.Buffer
	if err := advancedAnalyticsTmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		h.log.Error("advanced analytics: render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w) //nolint:errcheck
}
