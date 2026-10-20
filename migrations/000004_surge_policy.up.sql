CREATE TABLE surge_policies (
    id UUID PRIMARY KEY,
    locality_id UUID NOT NULL REFERENCES localities(id),
    version BIGINT NOT NULL CHECK (version > 0),
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE')),
    min_multiplier DOUBLE PRECISION NOT NULL DEFAULT 1.0 CHECK (min_multiplier >= 1.0),
    max_multiplier DOUBLE PRECISION NOT NULL DEFAULT 1.5 CHECK (max_multiplier >= min_multiplier AND max_multiplier <= 1.75),
    minimum_demand INTEGER NOT NULL DEFAULT 3 CHECK (minimum_demand >= 0),
    effective_from TIMESTAMPTZ NOT NULL,
    effective_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (effective_until IS NULL OR effective_until > effective_from),
    UNIQUE (locality_id, version)
);

CREATE INDEX idx_surge_policies_runtime
    ON surge_policies(locality_id, status, effective_from, effective_until);

CREATE TABLE surge_policy_tiers (
    policy_id UUID NOT NULL REFERENCES surge_policies(id) ON DELETE CASCADE,
    min_imbalance_ratio DOUBLE PRECISION NOT NULL CHECK (min_imbalance_ratio > 0),
    multiplier DOUBLE PRECISION NOT NULL CHECK (multiplier >= 1.0 AND multiplier <= 1.75),
    PRIMARY KEY (policy_id, min_imbalance_ratio)
);
