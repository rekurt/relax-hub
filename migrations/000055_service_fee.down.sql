ALTER TABLE bookings DROP COLUMN IF EXISTS service_fee_amount;

DROP TABLE IF EXISTS service_fee_configs;
