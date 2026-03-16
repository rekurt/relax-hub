-- Replace non-unique indexes with unique constraints for data integrity
DROP INDEX IF EXISTS idx_payments_booking_id;
DROP INDEX IF EXISTS idx_payments_external_id;

CREATE UNIQUE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE UNIQUE INDEX idx_payments_external_id ON payments(external_id) WHERE external_id IS NOT NULL AND external_id != '';
