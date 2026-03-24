ALTER TABLE bathhouses
    ADD COLUMN buffer_minutes INT NOT NULL DEFAULT 30,
    ADD COLUMN lead_time_hours INT NOT NULL DEFAULT 2,
    ADD COLUMN max_advance_days INT NOT NULL DEFAULT 90;

ALTER TABLE bathhouses
    ADD CONSTRAINT chk_buffer_minutes CHECK (buffer_minutes BETWEEN 0 AND 120),
    ADD CONSTRAINT chk_lead_time_hours CHECK (lead_time_hours BETWEEN 0 AND 48),
    ADD CONSTRAINT chk_max_advance_days CHECK (max_advance_days BETWEEN 7 AND 365);
