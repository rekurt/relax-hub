DROP TABLE IF EXISTS force_majeure_events;

-- Revert bookings.status width
ALTER TABLE bookings ALTER COLUMN status TYPE VARCHAR(20);
