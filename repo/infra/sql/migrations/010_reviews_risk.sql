DO $$ BEGIN
    CREATE TYPE credit_tier AS ENUM ('bronze', 'silver', 'gold', 'platinum', 'restricted');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS reviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reviewer_id     UUID NOT NULL REFERENCES users(id),
    subject_id      UUID NOT NULL REFERENCES users(id),
    order_type      VARCHAR NOT NULL,
    order_id        UUID NOT NULL,
    overall_rating  NUMERIC(3,1) NOT NULL,
    comment         TEXT,
    editable_until  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (reviewer_id, order_type, order_id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_subject ON reviews (subject_id);
CREATE INDEX IF NOT EXISTS idx_reviews_order ON reviews (order_type, order_id);

CREATE TABLE IF NOT EXISTS review_dimensions (
    id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name   VARCHAR NOT NULL UNIQUE,
    label  VARCHAR NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS review_scores (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id    UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    dimension_id UUID NOT NULL REFERENCES review_dimensions(id),
    score        NUMERIC(3,1) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_review_scores_review ON review_scores (review_id);

CREATE TABLE IF NOT EXISTS credit_tiers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_name        credit_tier NOT NULL,
    min_transactions INT NOT NULL,
    min_avg_rating   NUMERIC(3,1) NOT NULL,
    max_violations   INT NOT NULL,
    description      TEXT
);

CREATE TABLE IF NOT EXISTS user_credit_snapshots (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id),
    tier              credit_tier NOT NULL,
    total_transactions INT NOT NULL,
    avg_rating        NUMERIC(3,1) NOT NULL,
    violation_count   INT NOT NULL,
    computed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_credit_snapshots_user
    ON user_credit_snapshots (user_id, computed_at);

CREATE TABLE IF NOT EXISTS violation_records (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id),
    violation_type VARCHAR NOT NULL,
    description    TEXT,
    severity       VARCHAR NOT NULL,
    recorded_by    UUID NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_violation_records_user ON violation_records (user_id);

CREATE TABLE IF NOT EXISTS harassment_flags (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id      UUID NOT NULL REFERENCES users(id),
    subject_id       UUID NOT NULL REFERENCES users(id),
    description      TEXT NOT NULL,
    evidence_file_id UUID,
    status           VARCHAR NOT NULL DEFAULT 'open',
    reviewed_by      UUID,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_harassment_flags_subject ON harassment_flags (subject_id);
CREATE INDEX IF NOT EXISTS idx_harassment_flags_status ON harassment_flags (status);

CREATE TABLE IF NOT EXISTS no_show_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    order_type  VARCHAR NOT NULL,
    order_id    UUID NOT NULL,
    recorded_by UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_no_show_records_user ON no_show_records (user_id);

CREATE TABLE IF NOT EXISTS blacklist_records (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id),
    reason         TEXT,
    blacklisted_by UUID NOT NULL REFERENCES users(id),
    active         BOOLEAN NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lifted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_blacklist_records_user_active
    ON blacklist_records (user_id) WHERE active = true;

CREATE TABLE IF NOT EXISTS risk_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id),
    event_type    VARCHAR NOT NULL,
    description   TEXT,
    severity      VARCHAR NOT NULL,
    metadata_json JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_risk_events_user ON risk_events (user_id);
CREATE INDEX IF NOT EXISTS idx_risk_events_type ON risk_events (event_type);

CREATE TABLE IF NOT EXISTS risk_scores (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id),
    score        NUMERIC(5,2) NOT NULL,
    factors_json JSONB,
    computed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_risk_scores_user ON risk_scores (user_id, computed_at);

CREATE TABLE IF NOT EXISTS throttle_actions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    action_type VARCHAR NOT NULL,
    reason      TEXT,
    expires_at  TIMESTAMPTZ,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_throttle_actions_user_active
    ON throttle_actions (user_id) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_throttle_actions_expires ON throttle_actions (expires_at);

CREATE TABLE IF NOT EXISTS admin_approvals (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID,
    action_type      VARCHAR NOT NULL,
    reference_type   VARCHAR NOT NULL,
    reference_id     UUID NOT NULL,
    status           VARCHAR NOT NULL DEFAULT 'pending',
    requested_by     UUID,
    resolved_by      UUID,
    resolution_notes TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_admin_approvals_status ON admin_approvals (status);
CREATE INDEX IF NOT EXISTS idx_admin_approvals_reference
    ON admin_approvals (reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_admin_approvals_user ON admin_approvals (user_id);
