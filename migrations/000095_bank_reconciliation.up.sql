-- Bank statement entries for reconciliation
CREATE TABLE bank_statement_entries (
    id UUID PRIMARY KEY,
    date DATE NOT NULL,
    amount BIGINT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    counterparty TEXT NOT NULL DEFAULT '',
    reference_num TEXT NOT NULL DEFAULT '',
    matched_tx_id UUID,
    matched_tx_type VARCHAR(30) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    upload_batch_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bank_statement_entries_status ON bank_statement_entries(status) WHERE status = 'pending';
CREATE INDEX idx_bank_statement_entries_date ON bank_statement_entries(date);
CREATE INDEX idx_bank_statement_entries_batch ON bank_statement_entries(upload_batch_id);
CREATE INDEX idx_bank_statement_entries_amount ON bank_statement_entries(amount);

-- Bank statement uploads metadata
CREATE TABLE bank_statement_uploads (
    id UUID PRIMARY KEY,
    file_name TEXT NOT NULL,
    format VARCHAR(10) NOT NULL,
    total_rows INT NOT NULL DEFAULT 0,
    matched_count INT NOT NULL DEFAULT 0,
    pending_count INT NOT NULL DEFAULT 0,
    ignored_count INT NOT NULL DEFAULT 0,
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
