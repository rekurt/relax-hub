-- Owner metrics: response rate, avg response time, cancelled_by_owner tracking
ALTER TABLE bathhouses ADD COLUMN response_rate NUMERIC(5,4) NOT NULL DEFAULT 1.0;
ALTER TABLE bathhouses ADD COLUMN avg_response_time_minutes INT NOT NULL DEFAULT 0;

ALTER TABLE bookings ADD COLUMN cancelled_by_owner BOOLEAN NOT NULL DEFAULT false;
