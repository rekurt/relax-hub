-- Remove telegram notification preference column
ALTER TABLE notification_preferences DROP COLUMN IF EXISTS telegram;
