DROP TABLE IF EXISTS audit_logs;
DROP TRIGGER IF EXISTS update_site_config_updated_at ON site_config;
DROP TABLE IF EXISTS site_config;
DROP TRIGGER IF EXISTS update_permissions_updated_at ON permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_channel;
