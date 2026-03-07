-- GoAdmin system tables
-- Required by github.com/GoAdminGroup/go-admin for internal session, RBAC, menu, and audit functionality.

-- Sequences
CREATE SEQUENCE IF NOT EXISTS goadmin_users_myid_seq;
CREATE SEQUENCE IF NOT EXISTS goadmin_roles_myid_seq;
CREATE SEQUENCE IF NOT EXISTS goadmin_permissions_myid_seq;
CREATE SEQUENCE IF NOT EXISTS goadmin_menu_myid_seq;
CREATE SEQUENCE IF NOT EXISTS goadmin_operation_log_myid_seq;
CREATE SEQUENCE IF NOT EXISTS goadmin_session_myid_seq;
CREATE SEQUENCE IF NOT EXISTS goadmin_site_myid_seq;

-- Users (GoAdmin internal accounts, not the app users table)
CREATE TABLE goadmin_users (
    id          integer DEFAULT nextval('goadmin_users_myid_seq') NOT NULL PRIMARY KEY,
    username    varchar(190) NOT NULL UNIQUE,
    password    varchar(80)  NOT NULL,
    name        varchar(255) NOT NULL,
    avatar      varchar(255),
    remember_token varchar(100),
    created_at  timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at  timestamp WITHOUT TIME ZONE DEFAULT now()
);

-- Roles
CREATE TABLE goadmin_roles (
    id         integer DEFAULT nextval('goadmin_roles_myid_seq') NOT NULL PRIMARY KEY,
    name       varchar(255) NOT NULL UNIQUE,
    slug       varchar(255) NOT NULL,
    created_at timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at timestamp WITHOUT TIME ZONE DEFAULT now()
);

-- Permissions
CREATE TABLE goadmin_permissions (
    id          integer DEFAULT nextval('goadmin_permissions_myid_seq') NOT NULL PRIMARY KEY,
    name        varchar(50) NOT NULL UNIQUE,
    slug        varchar(50) NOT NULL,
    http_method varchar(255),
    http_path   text NOT NULL,
    created_at  timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at  timestamp WITHOUT TIME ZONE DEFAULT now()
);

-- Role <-> User junction
CREATE TABLE goadmin_role_users (
    role_id    integer NOT NULL,
    user_id    integer NOT NULL,
    created_at timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at timestamp WITHOUT TIME ZONE DEFAULT now(),
    UNIQUE (role_id, user_id)
);

-- User <-> Permission junction
CREATE TABLE goadmin_user_permissions (
    user_id       integer NOT NULL,
    permission_id integer NOT NULL,
    created_at    timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at    timestamp WITHOUT TIME ZONE DEFAULT now(),
    UNIQUE (user_id, permission_id)
);

-- Role <-> Permission junction
CREATE TABLE goadmin_role_permissions (
    role_id       integer NOT NULL,
    permission_id integer NOT NULL,
    created_at    timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at    timestamp WITHOUT TIME ZONE DEFAULT now(),
    UNIQUE (role_id, permission_id)
);

-- Menu
CREATE TABLE goadmin_menu (
    id          integer DEFAULT nextval('goadmin_menu_myid_seq') NOT NULL PRIMARY KEY,
    parent_id   integer DEFAULT 0 NOT NULL,
    type        integer DEFAULT 0,
    "order"     integer DEFAULT 0 NOT NULL,
    title       varchar(50) NOT NULL,
    header      varchar(100),
    icon        varchar(50) NOT NULL,
    uri         varchar(50) NOT NULL,
    uuid        varchar(100),
    plugin_name varchar(150) NOT NULL DEFAULT '',
    created_at  timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at  timestamp WITHOUT TIME ZONE DEFAULT now()
);

-- Role <-> Menu junction
CREATE TABLE goadmin_role_menu (
    role_id    integer NOT NULL,
    menu_id    integer NOT NULL,
    created_at timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at timestamp WITHOUT TIME ZONE DEFAULT now()
);
CREATE INDEX idx_goadmin_role_menu_role_menu ON goadmin_role_menu (role_id, menu_id);

-- Operation log (audit trail)
CREATE TABLE goadmin_operation_log (
    id         integer DEFAULT nextval('goadmin_operation_log_myid_seq') NOT NULL PRIMARY KEY,
    user_id    integer NOT NULL,
    path       varchar(255) NOT NULL,
    method     varchar(10)  NOT NULL,
    ip         varchar(15)  NOT NULL,
    input      text         NOT NULL,
    created_at timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at timestamp WITHOUT TIME ZONE DEFAULT now()
);
CREATE INDEX idx_goadmin_operation_log_user_id ON goadmin_operation_log (user_id);

-- Session
CREATE TABLE goadmin_session (
    id         integer DEFAULT nextval('goadmin_session_myid_seq') NOT NULL PRIMARY KEY,
    sid        varchar(50)   NOT NULL,
    "values"   varchar(3000) NOT NULL,
    created_at timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at timestamp WITHOUT TIME ZONE DEFAULT now()
);

-- Site configuration
CREATE TABLE goadmin_site (
    id          integer DEFAULT nextval('goadmin_site_myid_seq') NOT NULL PRIMARY KEY,
    key         varchar(100) NOT NULL,
    value       text         NOT NULL,
    type        integer DEFAULT 0,
    description varchar(3000),
    state       integer DEFAULT 0,
    created_at  timestamp WITHOUT TIME ZONE DEFAULT now(),
    updated_at  timestamp WITHOUT TIME ZONE DEFAULT now()
);

-- Seed: default administrator role
INSERT INTO goadmin_roles (name, slug) VALUES ('Administrator', 'administrator');

-- Seed: wildcard permission for admin role
INSERT INTO goadmin_permissions (name, slug, http_method, http_path)
VALUES ('All', '*', '', '*');

-- Bind admin role to wildcard permission
INSERT INTO goadmin_role_permissions (role_id, permission_id)
VALUES (
    (SELECT id FROM goadmin_roles WHERE slug = 'administrator'),
    (SELECT id FROM goadmin_permissions WHERE slug = '*')
);

-- Seed: default admin menu items (GoAdmin built-in pages)
INSERT INTO goadmin_menu (parent_id, type, "order", title, icon, uri, header, plugin_name) VALUES
    (0, 1, 0, 'Admin',       'fa-blind', '',              '', ''),
    (1, 1, 0, 'Users',       'fa-users', '/info/manager', '', ''),
    (1, 1, 0, 'Roles',       'fa-user',  '/info/roles',   '', ''),
    (1, 1, 0, 'Permissions', 'fa-ban',   '/info/permission','','');

-- Bridge: create GoAdmin admin user for each existing app admin.
-- Uses bcrypt hash of "admin" as temporary password ($2a$10$... is a well-known bcrypt hash).
-- Admins should change the password via the GoAdmin UI after first login.
-- Password: "admin" hashed with bcrypt cost 10.
INSERT INTO goadmin_users (username, password, name, avatar, created_at, updated_at)
SELECT
    u.email,
    '$2a$10$YVwBMSaIkRl3W6V2jPBs5OKnNHCzz6YMCGhIXI3.LbEMYP5lSP4xO',
    u.name,
    '',
    u.created_at,
    u.updated_at
FROM users u
WHERE u.role = 'admin' AND u.is_active = true
ON CONFLICT (username) DO NOTHING;

-- Assign admin role to all bridged GoAdmin users
INSERT INTO goadmin_role_users (role_id, user_id)
SELECT
    (SELECT id FROM goadmin_roles WHERE slug = 'administrator'),
    ga.id
FROM goadmin_users ga
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_role_users ru WHERE ru.user_id = ga.id
);
