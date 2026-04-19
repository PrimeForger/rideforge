CREATE TABLE IF NOT EXISTS rides (
    id UUID PRIMARY KEY,
    rider_id UUID NOT NULL,
    driver_id UUID,
    status VARCHAR(20) NOT NULL,
    version INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY,
    reserved_for_ride UUID NULL,
    reserved_at TIMESTAMP NULL,
    status VARCHAR(20) NOT NULL
);

CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_rides_id_version ON rides(id, version);
CREATE INDEX idx_drivers_status ON drivers(status);

CREATE TABLE ride_driver_offers (
    ride_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL, -- OFFERED, REJECTED, ACCEPTED
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