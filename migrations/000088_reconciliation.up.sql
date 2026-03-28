-- Float snapshots: ежедневный снимок баланса платформы
CREATE TABLE float_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_wallets_total BIGINT NOT NULL DEFAULT 0,
    owner_wallets_total BIGINT NOT NULL DEFAULT 0,
    escrow_held_total BIGINT NOT NULL DEFAULT 0,
    wallet_holds_total BIGINT NOT NULL DEFAULT 0,
    expected_total BIGINT NOT NULL DEFAULT 0,
    actual_total BIGINT NOT NULL DEFAULT 0,
    discrepancy BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'ok',
    client_wallets_count INT NOT NULL DEFAULT 0,
    owner_wallets_count INT NOT NULL DEFAULT 0,
    escrow_count INT NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    snapshot_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_float_snapshots_date ON float_snapshots(snapshot_date);
CREATE INDEX idx_float_snapshots_status ON float_snapshots(status) WHERE status = 'discrepancy';

-- Reconciliation reports: сверка с платёжным провайдером
CREATE TABLE reconciliation_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    internal_payments_sum BIGINT NOT NULL DEFAULT 0,
    internal_payments_count INT NOT NULL DEFAULT 0,
    internal_refunds_sum BIGINT NOT NULL DEFAULT 0,
    internal_refunds_count INT NOT NULL DEFAULT 0,
    provider_payments_sum BIGINT NOT NULL DEFAULT 0,
    provider_payments_count INT NOT NULL DEFAULT 0,
    provider_refunds_sum BIGINT NOT NULL DEFAULT 0,
    provider_refunds_count INT NOT NULL DEFAULT 0,
    payment_discrepancy BIGINT NOT NULL DEFAULT 0,
    refund_discrepancy BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    mismatch_details TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reconciliation_reports_period ON reconciliation_reports(period_start, period_end);
CREATE INDEX idx_reconciliation_reports_status ON reconciliation_reports(status) WHERE status IN ('mismatch', 'error');
