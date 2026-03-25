ALTER TABLE reviews
    ADD COLUMN cleanliness DECIMAL(2,1),
    ADD COLUMN accuracy DECIMAL(2,1),
    ADD COLUMN communication DECIMAL(2,1),
    ADD COLUMN value_for_money DECIMAL(2,1);

-- Add check constraints for valid ranges (1.0-5.0, step 0.5)
ALTER TABLE reviews
    ADD CONSTRAINT chk_cleanliness CHECK (cleanliness IS NULL OR (cleanliness >= 1.0 AND cleanliness <= 5.0 AND MOD(cleanliness * 2, 1) = 0)),
    ADD CONSTRAINT chk_accuracy CHECK (accuracy IS NULL OR (accuracy >= 1.0 AND accuracy <= 5.0 AND MOD(accuracy * 2, 1) = 0)),
    ADD CONSTRAINT chk_communication CHECK (communication IS NULL OR (communication >= 1.0 AND communication <= 5.0 AND MOD(communication * 2, 1) = 0)),
    ADD CONSTRAINT chk_value_for_money CHECK (value_for_money IS NULL OR (value_for_money >= 1.0 AND value_for_money <= 5.0 AND MOD(value_for_money * 2, 1) = 0));
