DROP INDEX IF EXISTS idx_slot_blocks_calendar_id;
ALTER TABLE slot_blocks DROP COLUMN IF EXISTS external_calendar_id;
