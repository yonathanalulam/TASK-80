CREATE TABLE IF NOT EXISTS message_templates (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_key     VARCHAR NOT NULL UNIQUE,
    subject_template TEXT,
    body_template    TEXT,
    channel_type     VARCHAR NOT NULL DEFAULT 'in_app',
    active           BOOLEAN NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messages (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id     UUID,
    recipient_id  UUID NOT NULL REFERENCES users(id),
    subject       VARCHAR,
    body          TEXT,
    template_id   UUID,
    metadata_json JSONB,
    read_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient ON messages (recipient_id, read_at);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at);

CREATE TABLE IF NOT EXISTS notification_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type   VARCHAR NOT NULL,
    source_type  VARCHAR NOT NULL,
    source_id    UUID,
    payload_json JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_events_type ON notification_events (event_type);
CREATE INDEX IF NOT EXISTS idx_notification_events_source
    ON notification_events (source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_notification_events_created_at ON notification_events (created_at);

CREATE TABLE IF NOT EXISTS notification_recipients (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id       UUID NOT NULL REFERENCES notification_events(id),
    user_id        UUID NOT NULL REFERENCES users(id),
    channel        VARCHAR NOT NULL DEFAULT 'in_app',
    status         VARCHAR NOT NULL DEFAULT 'pending',
    delivered_at   TIMESTAMPTZ,
    read_at        TIMESTAMPTZ,
    deferred_until TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_recipients_event ON notification_recipients (event_id);
CREATE INDEX IF NOT EXISTS idx_notification_recipients_user
    ON notification_recipients (user_id, status);
CREATE INDEX IF NOT EXISTS idx_notification_recipients_deferred
    ON notification_recipients (deferred_until) WHERE deferred_until IS NOT NULL;

CREATE TABLE IF NOT EXISTS callback_queue_entries (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id         UUID REFERENCES notification_events(id),
    recipient_id     UUID,
    payload_json     JSONB,
    status           VARCHAR NOT NULL DEFAULT 'pending',
    attempts         INT NOT NULL DEFAULT 0,
    last_attempted_at TIMESTAMPTZ,
    exported_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_callback_queue_status ON callback_queue_entries (status);
CREATE INDEX IF NOT EXISTS idx_callback_queue_event ON callback_queue_entries (event_id);

CREATE TABLE IF NOT EXISTS send_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_user_id   UUID NOT NULL REFERENCES users(id),
    message_id          UUID,
    event_type          VARCHAR NOT NULL,
    channel_type        VARCHAR NOT NULL DEFAULT 'in_app',
    status              VARCHAR NOT NULL,
    payload_summary_json JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_send_logs_recipient ON send_logs (recipient_user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_send_logs_event_type ON send_logs (event_type);

CREATE TABLE IF NOT EXISTS message_read_receipts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES messages(id),
    user_id    UUID NOT NULL REFERENCES users(id),
    read_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_message_read_receipts_message ON message_read_receipts (message_id);
CREATE INDEX IF NOT EXISTS idx_message_read_receipts_user ON message_read_receipts (user_id);
