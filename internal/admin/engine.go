package admin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/admin/pages"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// PagesRouter creates a chi router with all custom admin pages.
// Dashboard is mounted at "/" as the default landing page.
func PagesRouter(
	dashProvider pages.DashboardDataProvider,
	modProvider pages.ModerationDataProvider,
	analyticsProvider pages.AnalyticsDataProvider,
	healthProvider pages.HealthDataProvider,
	log *logger.Logger,
	adminPrefix string,
) http.Handler {
	r := chi.NewRouter()

	pagesPrefix := adminPrefix + "/pages"

	dashboard := pages.NewDashboardHandler(dashProvider, log, adminPrefix)
	r.Get("/", dashboard.ServeHTTP)
	r.Get("/dashboard", dashboard.ServeHTTP)

	moderation := pages.NewModerationHandler(modProvider, log, pagesPrefix)
	r.Get("/moderation", moderation.ServeHTTP)
	r.Post("/moderation/api/approve", moderation.HandleApprove)
	r.Post("/moderation/api/reject", moderation.HandleReject)
	r.Post("/moderation/api/batch-approve", moderation.HandleBatchApprove)
	r.Post("/moderation/api/batch-reject", moderation.HandleBatchReject)

	analytics := pages.NewAnalyticsHandler(analyticsProvider, log, pagesPrefix)
	r.Get("/analytics", analytics.ServeHTTP)

	health := pages.NewHealthHandler(healthProvider, log)
	r.Get("/health", health.ServeHTTP)

	return r
}

// CustomMenuItem represents a menu entry to register with GoAdmin.
type CustomMenuItem struct {
	Title    string
	Icon     string
	URI      string
	Order    int
	ParentID int
	Header   string
}

// CustomMenuConfig returns the menu structure for all custom admin pages.
// Dashboard is the first item with a home icon.
// "Operations" is a group containing Moderation, Analytics, and Health.
func CustomMenuConfig(pagesPrefix string) []CustomMenuItem {
	return []CustomMenuItem{
		{
			Title: "Дашборд",
			Icon:  "fa-dashboard",
			URI:   pagesPrefix + "/",
			Order: 1,
		},
		{
			Title:  "Операции",
			Icon:   "fa-cogs",
			URI:    "",
			Order:  50,
			Header: "Операции",
		},
		{
			Title: "Модерация отзывов",
			Icon:  "fa-gavel",
			URI:   pagesPrefix + "/moderation",
			Order: 1,
		},
		{
			Title: "Аналитика",
			Icon:  "fa-line-chart",
			URI:   pagesPrefix + "/analytics",
			Order: 2,
		},
		{
			Title: "Мониторинг",
			Icon:  "fa-heartbeat",
			URI:   pagesPrefix + "/health",
			Order: 3,
		},
	}
}

// RegisterCustomMenu inserts custom page menu items into GoAdmin's goadmin_menu table.
// It is idempotent: existing custom menu items (identified by URI prefix) are removed first.
// The menu structure is: Dashboard (top-level), Operations group with Moderation/Analytics/Health children.
func RegisterCustomMenu(ctx context.Context, pool *pgxpool.Pool, pagesPrefix string, log *logger.Logger) error {
	items := CustomMenuConfig(pagesPrefix)
	if len(items) < 5 {
		return fmt.Errorf("expected 5 menu items, got %d", len(items))
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	now := time.Now()

	// Clean up existing custom menu items (idempotent).
	_, err = tx.Exec(ctx, `DELETE FROM goadmin_menu WHERE uri LIKE $1 OR header = 'Операции'`, pagesPrefix+"%")
	if err != nil {
		log.Error("admin menu: cleanup failed", "error", err)
		return fmt.Errorf("cleanup menu items: %w", err)
	}

	// Insert Dashboard (top-level, first item).
	dash := items[0]
	_, err = tx.Exec(ctx,
		`INSERT INTO goadmin_menu (parent_id, type, "order", title, icon, uri, header, plugin_name, uuid, created_at, updated_at)
		 VALUES (0, 0, $1, $2, $3, $4, '', '', '', $5, $5)`,
		dash.Order, dash.Title, dash.Icon, dash.URI, now,
	)
	if err != nil {
		log.Error("admin menu: insert dashboard", "error", err)
		return fmt.Errorf("insert dashboard menu: %w", err)
	}

	// Insert Operations group (parent for moderation, analytics, health).
	ops := items[1]
	var opsID int
	err = tx.QueryRow(ctx,
		`INSERT INTO goadmin_menu (parent_id, type, "order", title, icon, uri, header, plugin_name, uuid, created_at, updated_at)
		 VALUES (0, 0, $1, $2, $3, '', $4, '', '', $5, $5) RETURNING id`,
		ops.Order, ops.Title, ops.Icon, ops.Header, now,
	).Scan(&opsID)
	if err != nil {
		log.Error("admin menu: insert operations group", "error", err)
		return fmt.Errorf("insert operations group: %w", err)
	}

	// Insert child items under Operations.
	for _, item := range items[2:] {
		_, err = tx.Exec(ctx,
			`INSERT INTO goadmin_menu (parent_id, type, "order", title, icon, uri, header, plugin_name, uuid, created_at, updated_at)
			 VALUES ($1, 0, $2, $3, $4, $5, '', '', '', $6, $6)`,
			opsID, item.Order, item.Title, item.Icon, item.URI, now,
		)
		if err != nil {
			log.Error("admin menu: insert child", "error", err, "title", item.Title)
			return fmt.Errorf("insert menu item %s: %w", item.Title, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit menu transaction: %w", err)
	}

	log.Info("admin menu: registered custom pages",
		"dashboard", dash.URI,
		"operations_children", len(items[2:]),
	)
	return nil
}
