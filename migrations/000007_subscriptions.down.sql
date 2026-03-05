DROP INDEX IF EXISTS idx_promotions_dates;
DROP INDEX IF EXISTS idx_promotions_target_city_id;
DROP INDEX IF EXISTS idx_promotions_status;
DROP INDEX IF EXISTS idx_promotions_bathhouse_id;
DROP TABLE IF EXISTS promotions;

DROP INDEX IF EXISTS idx_subscriptions_bathhouse_active;
DROP INDEX IF EXISTS idx_subscriptions_end_date;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_owner_id;
DROP INDEX IF EXISTS idx_subscriptions_bathhouse_id;
DROP TABLE IF EXISTS subscriptions;
