-- Increase goadmin_menu.uri column length to support longer admin prefixes.
-- Default prefix /admin-panel/pages/moderation (29 chars) fits in varchar(50),
-- but custom longer prefixes could exceed it.
ALTER TABLE goadmin_menu ALTER COLUMN uri TYPE varchar(255);
