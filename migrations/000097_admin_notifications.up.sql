CREATE TABLE admin_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role VARCHAR(30) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    data JSONB NOT NULL DEFAULT '{}',
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    read_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_admin_notifications_role ON admin_notifications (role);
CREATE INDEX idx_admin_notifications_role_read ON admin_notifications (role, is_read) WHERE is_read = false;
CREATE INDEX idx_admin_notifications_created_at ON admin_notifications (created_at DESC);
CREATE INDEX idx_admin_notifications_type ON admin_notifications (type);
CREATE INDEX idx_admin_notifications_severity ON admin_notifications (severity) WHERE severity IN ('error', 'critical');
