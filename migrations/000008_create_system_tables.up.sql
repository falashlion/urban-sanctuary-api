-- Notification channel enum
DO $$ BEGIN
    CREATE TYPE notification_channel AS ENUM ('email', 'sms', 'push');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       VARCHAR(100) NOT NULL,
    channel    notification_channel NOT NULL DEFAULT 'email',
    title      VARCHAR(255) NOT NULL,
    content    TEXT NOT NULL,
    metadata   JSONB,
    sent_at    TIMESTAMPTZ,
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read_at ON notifications(read_at) WHERE read_at IS NULL;

-- Permissions table (RBAC matrix)
CREATE TABLE IF NOT EXISTS permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role        user_role NOT NULL,
    module      VARCHAR(100) NOT NULL,
    can_read    BOOLEAN NOT NULL DEFAULT false,
    can_write   BOOLEAN NOT NULL DEFAULT false,
    can_delete  BOOLEAN NOT NULL DEFAULT false,
    can_approve BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT permissions_unique_role_module UNIQUE (role, module)
);

DROP TRIGGER IF EXISTS update_permissions_updated_at ON permissions;
CREATE TRIGGER update_permissions_updated_at
    BEFORE UPDATE ON permissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Site configuration table
CREATE TABLE IF NOT EXISTS site_config (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         VARCHAR(255) NOT NULL UNIQUE,
    value       TEXT NOT NULL,
    description TEXT,
    updated_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS update_site_config_updated_at ON site_config;
CREATE TRIGGER update_site_config_updated_at
    BEFORE UPDATE ON site_config
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    action        VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id   UUID,
    metadata      JSONB,
    ip_address    VARCHAR(45),
    user_agent    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Seed default permissions
INSERT INTO permissions (role, module, can_read, can_write, can_delete, can_approve) VALUES
    ('admin', 'users', true, true, true, true),
    ('admin', 'properties', true, true, true, true),
    ('admin', 'bookings', true, true, true, true),
    ('admin', 'payments', true, true, false, true),
    ('admin', 'reviews', true, true, true, true),
    ('admin', 'tickets', true, true, true, true),
    ('admin', 'config', true, true, false, false),
    ('admin', 'audit_logs', true, false, false, false),
    ('homeowner', 'properties', true, true, false, false),
    ('homeowner', 'bookings', true, false, false, false),
    ('homeowner', 'reviews', true, false, false, false),
    ('homeowner', 'tickets', true, true, false, false),
    ('guest', 'properties', true, false, false, false),
    ('guest', 'bookings', true, true, false, false),
    ('guest', 'reviews', true, true, false, false),
    ('guest', 'tickets', true, true, false, false)
ON CONFLICT (role, module) DO NOTHING;

-- Seed default site config
INSERT INTO site_config (key, value, description) VALUES
    ('site_name', 'Urban Sanctuary', 'The name of the platform'),
    ('site_currency', 'XAF', 'Primary currency (CFA Franc)'),
    ('loyalty_points_per_booking', '100', 'Points earned per completed booking'),
    ('loyalty_points_value_xaf', '10', 'XAF value per loyalty point when redeemed'),
    ('max_booking_advance_days', '365', 'Maximum days in advance a booking can be made'),
    ('min_booking_nights', '1', 'Minimum nights per booking'),
    ('max_booking_nights', '30', 'Maximum nights per booking'),
    ('cancellation_window_hours', '24', 'Hours before check-in to allow free cancellation'),
    ('terms_of_service', 'https://urbansanctuary.cm/terms', 'Terms of Service URL'),
    ('about_us', 'Urban Sanctuary is a premium apartment booking platform in Douala, Cameroon.', 'About Us text')
ON CONFLICT (key) DO NOTHING;
