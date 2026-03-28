-- Per-event-type notification preferences with SMS channel support.
-- Extends the existing category-based notification_preferences table
-- with fine-grained per-event control.

-- Add sms_enabled column to existing notification_preferences table
ALTER TABLE notification_preferences ADD COLUMN IF NOT EXISTS sms BOOLEAN NOT NULL DEFAULT false;

-- Per-event notification preferences: overrides category-level settings
CREATE TABLE IF NOT EXISTS notification_event_preferences (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    push_enabled  BOOLEAN NOT NULL DEFAULT true,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    sms_enabled   BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (user_id, event_type)
);

CREATE INDEX IF NOT EXISTS idx_notification_event_prefs_user ON notification_event_preferences(user_id);

-- Push delivery tracking for fallback chain
CREATE TABLE IF NOT EXISTS push_delivery_log (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status        VARCHAR(20) NOT NULL DEFAULT 'sent', -- sent, delivered, failed
    sent_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at  TIMESTAMPTZ,
    fallback_sent BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_push_delivery_log_notif ON push_delivery_log(notification_id);
CREATE INDEX IF NOT EXISTS idx_push_delivery_log_pending ON push_delivery_log(status, sent_at)
    WHERE status = 'sent' AND fallback_sent = false;
