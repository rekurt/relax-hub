-- Drop GoAdmin system tables (reverse of up migration)
DROP TABLE IF EXISTS goadmin_site;
DROP TABLE IF EXISTS goadmin_session;
DROP TABLE IF EXISTS goadmin_operation_log;
DROP TABLE IF EXISTS goadmin_role_menu;
DROP TABLE IF EXISTS goadmin_menu;
DROP TABLE IF EXISTS goadmin_role_permissions;
DROP TABLE IF EXISTS goadmin_user_permissions;
DROP TABLE IF EXISTS goadmin_role_users;
DROP TABLE IF EXISTS goadmin_permissions;
DROP TABLE IF EXISTS goadmin_roles;
DROP TABLE IF EXISTS goadmin_users;

DROP SEQUENCE IF EXISTS goadmin_site_myid_seq;
DROP SEQUENCE IF EXISTS goadmin_session_myid_seq;
DROP SEQUENCE IF EXISTS goadmin_operation_log_myid_seq;
DROP SEQUENCE IF EXISTS goadmin_menu_myid_seq;
DROP SEQUENCE IF EXISTS goadmin_permissions_myid_seq;
DROP SEQUENCE IF EXISTS goadmin_roles_myid_seq;
DROP SEQUENCE IF EXISTS goadmin_users_myid_seq;
