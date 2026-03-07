-- Add telegram notification preference column
ALTER TABLE notification_preferences ADD COLUMN telegram BOOLEAN NOT NULL DEFAULT true;
