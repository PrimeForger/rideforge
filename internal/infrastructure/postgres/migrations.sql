-- Reference snapshot of the schema after root migrations/000005.
-- Apply migrations/ with golang-migrate; this file is not an installation script.
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS rides (
    id UUID PRIMARY KEY,
    rider_id UUID NOT NULL,
    driver_id UUID,
    status VARCHAR(20) NOT NULL,
    version INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    pickup_latitude DOUBLE PRECISION NOT NULL,
    pickup_longitude DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY,
    reserved_for_ride UUID NULL,
    reserved_at TIMESTAMP NULL,
    status VARCHAR(20) NOT NULL,
    rating FLOAT DEFAULT 5.0,
    acceptance_rate FLOAT DEFAULT 1.0,
    cancellation_rate FLOAT DEFAULT 0.0,
    timeout_rate FLOAT DEFAULT 0.0,
    completed_rides INT DEFAULT 0,
    last_assigned_at TIMESTAMP NULL,
    lat DOUBLE PRECISION DEFAULT 0,
    lng DOUBLE PRECISION DEFAULT 0
);

CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_rides_id_version ON rides(id, version);
CREATE INDEX idx_drivers_status ON drivers(status);

CREATE TABLE ride_driver_offers (
    ride_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL, -- OFFERED, REJECTED, ACCEPTED, TIMEOUT
    attempt INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (ride_id, driver_id)
);

CREATE INDEX idx_ride_offers_ride ON ride_driver_offers(ride_id);

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMP NULL
);

CREATE INDEX idx_outbox_unpublished ON outbox_events(published);

CREATE TABLE IF NOT EXISTS processed_events (
    event_id UUID NOT NULL,
    consumer_name VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, consumer_name)
);

CREATE TABLE IF NOT EXISTS driver_push_tokens (
    driver_id UUID NOT NULL,
    device_id VARCHAR(150) NOT NULL,
    platform VARCHAR(20) NOT NULL, -- android, ios
    token TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (driver_id, device_id)
);

CREATE INDEX idx_driver_push_tokens_driver ON driver_push_tokens(driver_id);
CREATE INDEX idx_driver_push_tokens_enabled ON driver_push_tokens(driver_id, enabled);

CREATE TABLE IF NOT EXISTS driver_metrics (
    driver_id UUID PRIMARY KEY,

    offered_count BIGINT NOT NULL DEFAULT 0,
    acked_count BIGINT NOT NULL DEFAULT 0,
    accepted_count BIGINT NOT NULL DEFAULT 0,
    rejected_count BIGINT NOT NULL DEFAULT 0,
    timeout_count BIGINT NOT NULL DEFAULT 0,
    completed_rides BIGINT NOT NULL DEFAULT 0,
    cancelled_count BIGINT NOT NULL DEFAULT 0,

    acceptance_rate DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    rejection_rate DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    timeout_rate DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    cancellation_rate DOUBLE PRECISION NOT NULL DEFAULT 0.0,

    last_offered_at TIMESTAMP NULL,
    last_accepted_at TIMESTAMP NULL,
    last_completed_at TIMESTAMP NULL,

    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS driver_metric_events (
    event_id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS countries (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS administrative_regions (
    id UUID PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES countries(id),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (country_id, code),
    UNIQUE (country_id, name)
);

CREATE TABLE IF NOT EXISTS localities (
    id UUID PRIMARY KEY,
    administrative_region_id UUID NOT NULL REFERENCES administrative_regions(id),
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    boundary geometry(MultiPolygon, 4326) NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'ACTIVE', 'INACTIVE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (administrative_region_id, slug),
    CHECK (ST_IsValid(boundary)),
    CHECK (NOT ST_IsEmpty(boundary))
);

CREATE INDEX IF NOT EXISTS idx_localities_region_status ON localities(administrative_region_id, status);
CREATE INDEX IF NOT EXISTS idx_localities_boundary ON localities USING GIST(boundary);

CREATE TABLE IF NOT EXISTS locality_h3_cells (
    locality_id UUID NOT NULL REFERENCES localities(id),
    h3_cell TEXT NOT NULL,
    h3_resolution SMALLINT NOT NULL CHECK (h3_resolution BETWEEN 0 AND 15),
    derived_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (locality_id, h3_cell),
    CONSTRAINT uq_locality_h3_cell UNIQUE (h3_cell)
);

CREATE TABLE IF NOT EXISTS surge_policies (
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
    UNIQUE (locality_id, version),
    CONSTRAINT ex_surge_policies_active_effective_period
        EXCLUDE USING gist (
            locality_id WITH =,
            tstzrange(effective_from, effective_until, '[)') WITH &&
        )
        WHERE (status = 'ACTIVE')
);

CREATE INDEX IF NOT EXISTS idx_surge_policies_runtime
    ON surge_policies(locality_id, status, effective_from, effective_until);

CREATE TABLE IF NOT EXISTS surge_policy_tiers (
    policy_id UUID NOT NULL REFERENCES surge_policies(id) ON DELETE CASCADE,
    min_imbalance_ratio DOUBLE PRECISION NOT NULL CHECK (min_imbalance_ratio > 0),
    multiplier DOUBLE PRECISION NOT NULL CHECK (multiplier >= 1.0 AND multiplier <= 1.75),
    PRIMARY KEY (policy_id, min_imbalance_ratio)
);
