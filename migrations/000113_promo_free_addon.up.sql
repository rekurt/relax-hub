ALTER TABLE promo_codes ADD COLUMN target_addon_id UUID REFERENCES add_ons(id) ON DELETE SET NULL;
