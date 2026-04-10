DO $$ BEGIN
    CREATE TYPE discount_type AS ENUM ('threshold_fixed', 'percentage', 'new_user_gift');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS coupons (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                VARCHAR NOT NULL UNIQUE,
    name                VARCHAR NOT NULL,
    discount_type       discount_type NOT NULL,
    amount              NUMERIC(18,2),
    min_spend           NUMERIC(18,2),
    percent_off         NUMERIC(5,2),
    valid_from          TIMESTAMPTZ,
    valid_to            TIMESTAMPTZ,
    eligibility_json    JSONB NOT NULL DEFAULT '{}',
    stack_group         VARCHAR,
    exclusive           BOOLEAN NOT NULL DEFAULT false,
    usage_limit_total   INT,
    usage_limit_per_user INT,
    active              BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons (code);
CREATE INDEX IF NOT EXISTS idx_coupons_active_valid
    ON coupons (active, valid_from, valid_to);

CREATE TABLE IF NOT EXISTS coupon_packs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS coupon_pack_items (
    coupon_pack_id UUID NOT NULL REFERENCES coupon_packs(id) ON DELETE CASCADE,
    coupon_id      UUID NOT NULL REFERENCES coupons(id) ON DELETE CASCADE,
    PRIMARY KEY (coupon_pack_id, coupon_id)
);

CREATE TABLE IF NOT EXISTS coupon_redemptions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_id            UUID NOT NULL REFERENCES coupons(id),
    user_id              UUID NOT NULL REFERENCES users(id),
    booking_id           UUID,
    procurement_order_id UUID,
    redemption_scope_key VARCHAR NOT NULL,
    discount_amount      NUMERIC(18,2) NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (coupon_id, user_id, redemption_scope_key)
);

CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon ON coupon_redemptions (coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_user ON coupon_redemptions (user_id);

CREATE TABLE IF NOT EXISTS pricing_memberships (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id),
    tier       VARCHAR NOT NULL,
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to   TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pricing_memberships_user ON pricing_memberships (user_id);
CREATE INDEX IF NOT EXISTS idx_pricing_memberships_active
    ON pricing_memberships (user_id, valid_from, valid_to);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id           UUID NOT NULL REFERENCES users(id),
    key                VARCHAR NOT NULL,
    route              VARCHAR NOT NULL,
    request_hash       VARCHAR,
    response_code      INT,
    response_body_json JSONB,
    locked_at          TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at         TIMESTAMPTZ NOT NULL,
    UNIQUE (actor_id, route, key)
);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires ON idempotency_keys (expires_at);
