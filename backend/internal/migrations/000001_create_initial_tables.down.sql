-- Drop services table related triggers first
DROP TRIGGER IF EXISTS trigger_services_updated_at;

-- Drop indexes for services table
DROP INDEX IF EXISTS idx_services_deleted_at;
DROP INDEX IF EXISTS idx_services_service_type_id;
DROP INDEX IF EXISTS idx_services_status;
DROP INDEX IF EXISTS idx_services_name;

-- Rollback V1.0 Core Tables

DROP TABLE IF EXISTS bugs;
DROP TABLE IF EXISTS business_client_types;
DROP TABLE IF EXISTS business_service_types;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS client_versions;
DROP TABLE IF EXISTS client_types;
DROP TABLE IF EXISTS businesses;
DROP TABLE IF EXISTS service_instances;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS service_types;
DROP TABLE IF EXISTS server_assets;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS environments;
DROP TABLE IF EXISTS responsibility_group_responsibilities;
DROP TABLE IF EXISTS responsibility_groups;
DROP TABLE IF EXISTS responsibilities;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS responsibility_groups_responsibilities; 