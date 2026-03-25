CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id),
    initiator_id UUID NOT NULL REFERENCES users(id),
    respondent_id UUID NOT NULL REFERENCES users(id),
    reason VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'open',
    resolution VARCHAR(30),
    refund_amount BIGINT NOT NULL DEFAULT 0,
    compensation_amount BIGINT NOT NULL DEFAULT 0,
    mediator_id UUID REFERENCES users(id),
    mediator_notes TEXT NOT NULL DEFAULT '',
    appeal_deadline TIMESTAMPTZ,
    evidence_deadline TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_disputes_booking_id ON disputes(booking_id);
CREATE INDEX idx_disputes_initiator_id ON disputes(initiator_id);
CREATE INDEX idx_disputes_respondent_id ON disputes(respondent_id);
CREATE INDEX idx_disputes_status ON disputes(status);

CREATE TABLE dispute_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(30) NOT NULL,
    url TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_dispute_evidence_dispute_id ON dispute_evidence(dispute_id);
