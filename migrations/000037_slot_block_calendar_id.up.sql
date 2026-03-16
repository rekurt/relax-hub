ALTER TABLE slot_blocks ADD COLUMN external_calendar_id UUID REFERENCES external_calendars(id) ON DELETE CASCADE;
CREATE INDEX idx_slot_blocks_calendar_id ON slot_blocks(external_calendar_id) WHERE external_calendar_id IS NOT NULL;
