package pages

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/logger"
)

var financeFuncMap = template.FuncMap{
	"kopecksToRub": KopecksToRub,
	"formatDate":   FormatDate,
	"reconStatus":  ReconStatus,
}

var financeTmpl = ParsePageTemplate(financeFuncMap, "templates/finance.tmpl")

// KopecksToRub formats kopecks as rubles with 2 decimal places.
func KopecksToRub(kopecks int64) string {
	rubles := float64(kopecks) / 100.0
	return fmt.Sprintf("%.2f", rubles)
}

// FormatDate formats time as DD.MM.YYYY.
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("02.01.2006")
}

// ReconStatus maps reconciliation status to Russian.
func ReconStatus(status string) string {
	switch status {
	case "ok":
		return "OK"
	case "discrepancy":
		return "Расхождение"
	case "matched":
		return "Совпадает"
	case "mismatch":
		return "Расхождение"
	case "error":
		return "Ошибка"
	case "pending":
		return "Ожидает"
	default:
		return status
	}
}

// FinanceSnapshot represents a float snapshot for the template.
type FinanceSnapshot struct {
	ClientWalletsTotal int64
	ClientWalletsCount int
	OwnerWalletsTotal  int64
	OwnerWalletsCount  int
	EscrowHeldTotal    int64
	EscrowCount        int
	WalletHoldsTotal   int64
	PlatformTotal      int64
	SnapshotDate       time.Time
	Status             string
}

// FinanceReconciliation represents a reconciliation report for the template.
type FinanceReconciliation struct {
	PeriodStart           time.Time
	PeriodEnd             time.Time
	InternalPaymentsSum   int64
	InternalPaymentsCount int
	InternalRefundsSum    int64
	InternalRefundsCount  int
	ProviderPaymentsSum   int64
	ProviderPaymentsCount int
	PaymentDiscrepancy    int64
	RefundDiscrepancy     int64
	Status                string
	CreatedAt             time.Time
}

// FinanceData is the full data model for the finance dashboard page.
type FinanceData struct {
	CurrentFloat    FinanceSnapshot
	LastSnapshot    *FinanceSnapshot
	LastRecon       *FinanceReconciliation
	RecentSnapshots []FinanceSnapshot
	GeneratedAt     time.Time
	PagesPrefix     string
	AdminPrefix     string
	PageTitle       string
	ActivePage      string
}

// FinanceDataProvider fetches finance data.
type FinanceDataProvider interface {
	GetFinanceData(ctx context.Context) (*FinanceData, error)
}

// PostgresFinanceProvider fetches finance data from PostgreSQL.
type PostgresFinanceProvider struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewPostgresFinanceProvider creates a new PostgresFinanceProvider.
func NewPostgresFinanceProvider(pool *pgxpool.Pool, log *logger.Logger) *PostgresFinanceProvider {
	return &PostgresFinanceProvider{pool: pool, log: log}
}

func (p *PostgresFinanceProvider) GetFinanceData(ctx context.Context) (*FinanceData, error) {
	data := &FinanceData{GeneratedAt: time.Now()}

	// Current float: sum wallets by role
	if err := p.loadCurrentFloat(ctx, data); err != nil {
		p.log.Error("finance: load current float", "error", err)
	}

	// Last snapshot
	if err := p.loadLastSnapshot(ctx, data); err != nil {
		p.log.Error("finance: load last snapshot", "error", err)
	}

	// Last reconciliation
	if err := p.loadLastReconciliation(ctx, data); err != nil {
		p.log.Error("finance: load last recon", "error", err)
	}

	// Recent snapshots (last 7 days)
	if err := p.loadRecentSnapshots(ctx, data); err != nil {
		p.log.Error("finance: load recent snapshots", "error", err)
	}

	return data, nil
}

func (p *PostgresFinanceProvider) loadCurrentFloat(ctx context.Context, data *FinanceData) error {
	err := p.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(w.balance), 0), COUNT(w.id)
		FROM wallets w JOIN users u ON u.id = w.user_id
		WHERE u.role = 'client' AND w.status != 'archived'`,
	).Scan(&data.CurrentFloat.ClientWalletsTotal, &data.CurrentFloat.ClientWalletsCount)
	if err != nil {
		return err
	}

	err = p.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(w.balance), 0), COUNT(w.id)
		FROM wallets w JOIN users u ON u.id = w.user_id
		WHERE u.role = 'owner' AND w.status != 'archived'`,
	).Scan(&data.CurrentFloat.OwnerWalletsTotal, &data.CurrentFloat.OwnerWalletsCount)
	if err != nil {
		return err
	}

	err = p.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0), COUNT(id)
		FROM escrows WHERE status IN ('held', 'disputed')`,
	).Scan(&data.CurrentFloat.EscrowHeldTotal, &data.CurrentFloat.EscrowCount)
	if err != nil {
		return err
	}

	err = p.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM wallet_holds WHERE status = 'active'`,
	).Scan(&data.CurrentFloat.WalletHoldsTotal)
	if err != nil {
		return err
	}

	data.CurrentFloat.PlatformTotal = data.CurrentFloat.ClientWalletsTotal +
		data.CurrentFloat.OwnerWalletsTotal + data.CurrentFloat.EscrowHeldTotal
	data.CurrentFloat.Status = "ok"

	return nil
}

func (p *PostgresFinanceProvider) loadLastSnapshot(ctx context.Context, data *FinanceData) error {
	var s FinanceSnapshot
	err := p.pool.QueryRow(ctx, `
		SELECT client_wallets_total, client_wallets_count, owner_wallets_total, owner_wallets_count,
			escrow_held_total, escrow_count, wallet_holds_total, expected_total, snapshot_date, status
		FROM float_snapshots ORDER BY snapshot_date DESC LIMIT 1`,
	).Scan(&s.ClientWalletsTotal, &s.ClientWalletsCount, &s.OwnerWalletsTotal, &s.OwnerWalletsCount,
		&s.EscrowHeldTotal, &s.EscrowCount, &s.WalletHoldsTotal, &s.PlatformTotal, &s.SnapshotDate, &s.Status)
	if err != nil {
		return nil // no snapshots yet is ok
	}
	data.LastSnapshot = &s
	return nil
}

func (p *PostgresFinanceProvider) loadLastReconciliation(ctx context.Context, data *FinanceData) error {
	var r FinanceReconciliation
	err := p.pool.QueryRow(ctx, `
		SELECT period_start, period_end,
			internal_payments_sum, internal_payments_count,
			internal_refunds_sum, internal_refunds_count,
			provider_payments_sum, provider_payments_count,
			payment_discrepancy, refund_discrepancy,
			status, created_at
		FROM reconciliation_reports ORDER BY created_at DESC LIMIT 1`,
	).Scan(&r.PeriodStart, &r.PeriodEnd,
		&r.InternalPaymentsSum, &r.InternalPaymentsCount,
		&r.InternalRefundsSum, &r.InternalRefundsCount,
		&r.ProviderPaymentsSum, &r.ProviderPaymentsCount,
		&r.PaymentDiscrepancy, &r.RefundDiscrepancy,
		&r.Status, &r.CreatedAt)
	if err != nil {
		return nil // no reports yet is ok
	}
	data.LastRecon = &r
	return nil
}

func (p *PostgresFinanceProvider) loadRecentSnapshots(ctx context.Context, data *FinanceData) error {
	rows, err := p.pool.Query(ctx, `
		SELECT client_wallets_total, client_wallets_count, owner_wallets_total, owner_wallets_count,
			escrow_held_total, escrow_count, wallet_holds_total, expected_total, snapshot_date, status
		FROM float_snapshots
		WHERE snapshot_date >= $1
		ORDER BY snapshot_date DESC LIMIT 7`,
		time.Now().AddDate(0, 0, -7))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var s FinanceSnapshot
		if err := rows.Scan(&s.ClientWalletsTotal, &s.ClientWalletsCount, &s.OwnerWalletsTotal, &s.OwnerWalletsCount,
			&s.EscrowHeldTotal, &s.EscrowCount, &s.WalletHoldsTotal, &s.PlatformTotal, &s.SnapshotDate, &s.Status); err != nil {
			return err
		}
		data.RecentSnapshots = append(data.RecentSnapshots, s)
	}
	return nil
}

// FinanceHandler serves the finance dashboard page.
type FinanceHandler struct {
	provider    FinanceDataProvider
	log         *logger.Logger
	pagesPrefix string
	adminPrefix string
}

// NewFinanceHandler creates a new admin FinanceHandler.
func NewFinanceHandler(provider FinanceDataProvider, log *logger.Logger, pagesPrefix, adminPrefix string) *FinanceHandler {
	return &FinanceHandler{provider: provider, log: log, pagesPrefix: pagesPrefix, adminPrefix: adminPrefix}
}

// ServeHTTP renders the finance dashboard page.
func (h *FinanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h.provider.GetFinanceData(r.Context())
	if err != nil {
		h.log.Error("finance: get data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.PagesPrefix = h.pagesPrefix
	data.AdminPrefix = h.adminPrefix
	data.PageTitle = "Финансовый мониторинг"
	data.ActivePage = "finance"

	var buf bytes.Buffer
	if err := financeTmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		h.log.Error("finance: render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w) //nolint:errcheck
}
