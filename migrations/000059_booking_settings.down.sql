ALTER TABLE bathhouses
    DROP CONSTRAINT IF EXISTS chk_buffer_minutes,
    DROP CONSTRAINT IF EXISTS chk_lead_time_hours,
    DROP CONSTRAINT IF EXISTS chk_max_advance_days;

ALTER TABLE bathhouses
    DROP COLUMN IF EXISTS buffer_minutes,
    DROP COLUMN IF EXISTS lead_time_hours,
    DROP COLUMN IF EXISTS max_advance_days;
