-- Track when a bathhouse's response rate first dropped below 30%.
-- After 60 consecutive days below threshold, booking_mode is forced to 'instant'.
ALTER TABLE bathhouses ADD COLUMN IF NOT EXISTS low_response_rate_since TIMESTAMPTZ;
