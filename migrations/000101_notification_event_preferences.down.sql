DROP TABLE IF EXISTS push_delivery_log;
DROP TABLE IF EXISTS notification_event_preferences;
ALTER TABLE notification_preferences DROP COLUMN IF EXISTS sms;
