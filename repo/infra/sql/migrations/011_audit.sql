
CREATE TABLE IF NOT EXISTS audit_logs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id       UUID,
    action         VARCHAR NOT NULL,
    entity_type    VARCHAR NOT NULL,
    entity_id      UUID,
    before_summary JSONB,
    after_summary  JSONB,
    request_id     UUID,
    ip_address     VARCHAR,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor ON audit_logs (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_request ON audit_logs (request_id);

CREATE TABLE IF NOT EXISTS entity_versions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type    VARCHAR NOT NULL,
    entity_id      UUID NOT NULL,
    version_number INT NOT NULL,
    data_json      JSONB,
    changed_by     UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entity_versions_entity
    ON entity_versions (entity_type, entity_id, version_number);

CREATE TABLE IF NOT EXISTS domain_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type     VARCHAR NOT NULL,
    aggregate_type VARCHAR NOT NULL,
    aggregate_id   UUID NOT NULL,
    payload_json   JSONB,
    actor_id       UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_domain_events_aggregate
    ON domain_events (aggregate_type, aggregate_id);
CREATE INDEX IF NOT EXISTS idx_domain_events_type ON domain_events (event_type);
CREATE INDEX IF NOT EXISTS idx_domain_events_created_at ON domain_events (created_at);
CREATE INDEX IF NOT EXISTS idx_domain_events_actor ON domain_events (actor_id);
