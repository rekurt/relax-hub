DROP INDEX IF EXISTS idx_payments_booking_id;
DROP INDEX IF EXISTS idx_payments_external_id;

CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_external_id ON payments(external_id);
