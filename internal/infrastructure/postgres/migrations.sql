CREATE TABLE IF NOT EXISTS rides (
    id UUID PRIMARY KEY,
    rider_id UUID NOT NULL,
    driver_id UUID,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY,
    status VARCHAR(20) NOT NULL
);

CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_drivers_status ON drivers(status);
