CREATE TYPE kyc_status AS ENUM ('pending', 'approved', 'rejected', 'expired');
CREATE TYPE kyc_entity_type AS ENUM ('individual', 'sole_proprietor', 'self_employed', 'legal_entity');

CREATE TABLE kyc_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status kyc_status NOT NULL DEFAULT 'pending',
    entity_type kyc_entity_type NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    inn VARCHAR(12) NOT NULL DEFAULT '',
    ogrnip VARCHAR(15) NOT NULL DEFAULT '',
    company_name VARCHAR(255) NOT NULL DEFAULT '',
    document_urls TEXT[] NOT NULL DEFAULT '{}',
    rejection_reason TEXT NOT NULL DEFAULT '',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_applications_user_id ON kyc_applications(user_id);
CREATE INDEX idx_kyc_applications_status ON kyc_applications(status);
CREATE INDEX idx_kyc_applications_expires_at ON kyc_applications(expires_at) WHERE expires_at IS NOT NULL;
