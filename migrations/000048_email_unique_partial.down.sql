DROP INDEX IF EXISTS idx_users_email_unique;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
CREATE INDEX idx_users_email ON users (email);
