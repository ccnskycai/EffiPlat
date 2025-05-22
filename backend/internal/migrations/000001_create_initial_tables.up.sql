-- V1.0 Core Tables

-- users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    department TEXT,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active', -- e.g., 'active', 'inactive', 'pending'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Use appropriate timestamp type
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- REMOVED ON UPDATE for SQLite compatibility
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- roles table
CREATE TABLE roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_roles_name ON roles(name);

-- permissions table
CREATE TABLE permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resource TEXT NOT NULL, -- e.g., 'environment', 'service', 'bug'
    action TEXT NOT NULL, -- e.g., 'create', 'read', 'update', 'delete', 'assign'
    description TEXT,
    UNIQUE (resource, action)
);
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);

-- user_roles table
CREATE TABLE user_roles (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- role_permissions table
CREATE TABLE role_permissions (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- responsibilities table
CREATE TABLE responsibilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_responsibilities_name ON responsibilities(name);

-- Create new responsibility_groups table
CREATE TABLE responsibility_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_responsibility_groups_name ON responsibility_groups(name);

-- Create responsibility_group_responsibilities join table
CREATE TABLE responsibility_group_responsibilities (
    responsibility_group_id INTEGER NOT NULL,
    responsibility_id INTEGER NOT NULL,
    PRIMARY KEY (responsibility_group_id, responsibility_id),
    FOREIGN KEY (responsibility_group_id) REFERENCES responsibility_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (responsibility_id) REFERENCES responsibilities(id) ON DELETE CASCADE
);

CREATE INDEX idx_resp_group_resp_group_id ON responsibility_group_responsibilities(responsibility_group_id);
CREATE INDEX idx_resp_group_resp_resp_id ON responsibility_group_responsibilities(responsibility_id);

-- environments table
CREATE TABLE environments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE, -- e.g., 'prod', 'staging', 'dev'
    description TEXT,
    type TEXT, -- e.g., 'physical', 'cloud', 'hybrid'
    status TEXT NOT NULL DEFAULT 'active', -- e.g., 'active', 'maintenance', 'decommissioned'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Use appropriate timestamp type
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- REMOVED ON UPDATE for SQLite compatibility
);
CREATE INDEX idx_environments_code ON environments(code);
CREATE INDEX idx_environments_status ON environments(status);

-- assets table
CREATE TABLE IF NOT EXISTS assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hostname VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(100) NOT NULL UNIQUE,
    asset_type VARCHAR(50) NOT NULL DEFAULT 'other',
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    description TEXT,
    environment_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE SET NULL -- or ON DELETE CASCADE depending on desired behavior
);

CREATE INDEX IF NOT EXISTS idx_assets_hostname ON assets (hostname);
CREATE INDEX IF NOT EXISTS idx_assets_ip_address ON assets (ip_address);
CREATE INDEX IF NOT EXISTS idx_assets_asset_type ON assets (asset_type);
CREATE INDEX IF NOT EXISTS idx_assets_status ON assets (status);
CREATE INDEX IF NOT EXISTS idx_assets_environment_id ON assets (environment_id);
CREATE INDEX IF NOT EXISTS idx_assets_deleted_at ON assets (deleted_at);

-- server_assets table
CREATE TABLE server_assets (
    asset_id INTEGER PRIMARY KEY, -- References assets.id
    ip_address TEXT UNIQUE,
    os TEXT,
    hostname TEXT UNIQUE,
    spec TEXT, -- Consider JSON for CPU, RAM, Disk
    access_info TEXT, -- Store reference/location, NOT secrets
    mac_address TEXT UNIQUE,
    cpu_model TEXT,
    cpu_cores INTEGER,
    memory_gb INTEGER,
    disk_info TEXT, -- Consider JSON
    serial_number TEXT UNIQUE,
    physical_location TEXT,
    status TEXT, -- More specific status if needed, else inherit from assets
    owner_group_id INTEGER, -- FK to responsibility_groups
    is_virtual INTEGER DEFAULT 0, -- 0 for false, 1 for true
    virtualization_host_id INTEGER, -- FK to assets.id
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_group_id) REFERENCES responsibility_groups(id) ON DELETE SET NULL,
    FOREIGN KEY (virtualization_host_id) REFERENCES assets(id) ON DELETE SET NULL
);
CREATE INDEX idx_server_assets_ip ON server_assets(ip_address);
CREATE INDEX idx_server_assets_hostname ON server_assets(hostname);
CREATE INDEX idx_server_assets_owner ON server_assets(owner_group_id);
CREATE INDEX idx_server_assets_virt_host ON server_assets(virtualization_host_id);

-- service_types table
CREATE TABLE service_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE, -- e.g., 'API', 'Frontend', 'Database', 'Worker'
    description TEXT
);
CREATE INDEX idx_service_types_name ON service_types(name);

-- services table
CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    version VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    external_link VARCHAR(2048),
    service_type_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (service_type_id) REFERENCES service_types(id) ON DELETE RESTRICT
);

-- Trigger to update 'updated_at' timestamp on row update for services
CREATE TRIGGER IF NOT EXISTS trigger_services_updated_at
AFTER UPDATE ON services
FOR EACH ROW
BEGIN
    UPDATE services SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Indexes for services table (moved from 20231027000004_create_services_indexes.up.sql)
CREATE INDEX IF NOT EXISTS idx_services_name ON services(name);
CREATE INDEX IF NOT EXISTS idx_services_status ON services(status);
CREATE INDEX IF NOT EXISTS idx_services_service_type_id ON services(service_type_id);
CREATE INDEX IF NOT EXISTS idx_services_deleted_at ON services(deleted_at);

-- service_instances table
CREATE TABLE service_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER NOT NULL,
    environment_id INTEGER NOT NULL,
    server_asset_id INTEGER NOT NULL, -- FK to assets.id
    port INTEGER,
    status TEXT NOT NULL DEFAULT 'running', -- e.g., 'running', 'stopped', 'deploying', 'error'
    version TEXT,
    -- created_at/updated_at might not be needed here or handled differently
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Use appropriate timestamp type
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- REMOVED ON UPDATE for SQLite compatibility
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
    FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE,
    FOREIGN KEY (server_asset_id) REFERENCES assets(id) ON DELETE CASCADE
);
CREATE INDEX idx_svc_inst_service_id ON service_instances(service_id);
CREATE INDEX idx_svc_inst_env_id ON service_instances(environment_id);
CREATE INDEX idx_svc_inst_asset_id ON service_instances(server_asset_id);
CREATE INDEX idx_svc_inst_status ON service_instances(status);

-- businesses table
CREATE TABLE businesses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Use appropriate timestamp type
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- REMOVED ON UPDATE for SQLite compatibility
);
CREATE INDEX idx_businesses_name ON businesses(name);

-- client_types table
CREATE TABLE client_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE, -- e.g., 'mac', 'win', 'android', 'ios', 'web', 'wechat'
    description TEXT
);
CREATE INDEX idx_client_types_name ON client_types(name);

-- client_versions table
CREATE TABLE client_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_type_id INTEGER NOT NULL,
    version TEXT NOT NULL,
    description TEXT,
    release_date TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Use appropriate timestamp type
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- REMOVED ON UPDATE for SQLite compatibility
    FOREIGN KEY (client_type_id) REFERENCES client_types(id) ON DELETE RESTRICT,
    UNIQUE (client_type_id, version)
);
CREATE INDEX idx_client_versions_type_id ON client_versions(client_type_id);
CREATE INDEX idx_client_versions_version ON client_versions(version);

-- clients table
CREATE TABLE clients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_version_id INTEGER NOT NULL,
    client_type_id INTEGER,
    ip TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Use appropriate timestamp type
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- REMOVED ON UPDATE for SQLite compatibility
    FOREIGN KEY (client_version_id) REFERENCES client_versions(id) ON DELETE CASCADE,
    FOREIGN KEY (client_type_id) REFERENCES client_types(id) ON DELETE SET NULL
);
CREATE INDEX idx_clients_version_id ON clients(client_version_id);
CREATE INDEX idx_clients_type_id ON clients(client_type_id);
CREATE INDEX idx_clients_ip ON clients(ip);

-- business_service_types table
CREATE TABLE business_service_types (
    business_id INTEGER NOT NULL,
    service_type_id INTEGER NOT NULL,
    PRIMARY KEY (business_id, service_type_id),
    FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE,
    FOREIGN KEY (service_type_id) REFERENCES service_types(id) ON DELETE CASCADE
);
CREATE INDEX idx_biz_svc_types_biz_id ON business_service_types(business_id);
CREATE INDEX idx_biz_svc_types_type_id ON business_service_types(service_type_id);

-- business_client_types table
CREATE TABLE business_client_types (
    business_id INTEGER NOT NULL,
    client_version_id INTEGER NOT NULL,
    PRIMARY KEY (business_id, client_version_id),
    FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE,
    FOREIGN KEY (client_version_id) REFERENCES client_versions(id) ON DELETE CASCADE
);
CREATE INDEX idx_biz_client_types_biz_id ON business_client_types(business_id);
CREATE INDEX idx_biz_client_types_ver_id ON business_client_types(client_version_id);

-- bugs table
CREATE TABLE bugs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'new',
    priority TEXT DEFAULT 'medium',
    reporter_id INTEGER NOT NULL,
    assignee_group_id INTEGER,
    environment_id INTEGER,
    service_instance_id INTEGER,
    business_id INTEGER,
    -- client_version_id INTEGER, -- Optional: If related to a specific client version
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Use appropriate timestamp type
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- REMOVED ON UPDATE for SQLite compatibility
    FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (assignee_group_id) REFERENCES responsibility_groups(id) ON DELETE SET NULL,
    FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE SET NULL,
    FOREIGN KEY (service_instance_id) REFERENCES service_instances(id) ON DELETE SET NULL,
    FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE SET NULL
    -- FOREIGN KEY (client_version_id) REFERENCES client_versions(id) ON DELETE SET NULL
);
CREATE INDEX idx_bugs_status ON bugs(status);
CREATE INDEX idx_bugs_priority ON bugs(priority);
CREATE INDEX idx_bugs_reporter_id ON bugs(reporter_id);
CREATE INDEX idx_bugs_assignee_group_id ON bugs(assignee_group_id);
CREATE INDEX idx_bugs_env_id ON bugs(environment_id);
CREATE INDEX idx_bugs_svc_inst_id ON bugs(service_instance_id);
CREATE INDEX idx_bugs_biz_id ON bugs(business_id);

-- COMMIT; -- Removed explicit COMMIT statement 