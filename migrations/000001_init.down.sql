DROP TRIGGER IF EXISTS trg_bathhouse_location ON bathhouses;
DROP FUNCTION IF EXISTS update_bathhouse_location();

DROP TABLE IF EXISTS representatives;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS bathhouses;
DROP TABLE IF EXISTS cities;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS postgis;
