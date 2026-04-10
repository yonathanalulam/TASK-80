DO $$ BEGIN
    CREATE TYPE user_status AS ENUM ('active', 'suspended', 'banned');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR NOT NULL,
    password_hash   VARCHAR NOT NULL,
    status          user_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email_active
    ON users (email) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at);

CREATE TABLE IF NOT EXISTS roles (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR NOT NULL UNIQUE
);

INSERT INTO roles (name) VALUES
    ('traveler'),
    ('group_organizer'),
    ('supplier'),
    ('courier_runner'),
    ('accountant'),
    ('administrator')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource    VARCHAR NOT NULL,
    action      VARCHAR NOT NULL,
    description TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_permissions_resource_action
    ON permissions (resource, action);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions (permission_id);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles (role_id);

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id                          UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name                     VARCHAR,
    phone_masked                     VARCHAR,
    emergency_contact_name_encrypted TEXT,
    emergency_contact_phone_encrypted TEXT,
    created_at                       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_preferences (
    user_id          UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    preferences_json JSONB NOT NULL DEFAULT '{}',
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS do_not_disturb_settings (
    user_id   UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    dnd_start TIME NOT NULL DEFAULT '21:00',
    dnd_end   TIME NOT NULL DEFAULT '08:00',
    enabled   BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS subscription_preferences (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_type VARCHAR NOT NULL,
    event_type   VARCHAR NOT NULL,
    enabled      BOOLEAN NOT NULL DEFAULT true,
    UNIQUE (user_id, channel_type, event_type)
);

CREATE INDEX IF NOT EXISTS idx_subscription_preferences_user ON subscription_preferences (user_id);

CREATE TABLE IF NOT EXISTS sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions (expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions (token_hash);
