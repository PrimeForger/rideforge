CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE countries (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE administrative_regions (
    id UUID PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES countries(id),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (country_id, code),
    UNIQUE (country_id, name)
);

CREATE TABLE localities (
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

CREATE INDEX idx_localities_region_status ON localities(administrative_region_id, status);
CREATE INDEX idx_localities_boundary ON localities USING GIST(boundary);

CREATE TABLE locality_h3_cells (
    locality_id UUID NOT NULL REFERENCES localities(id),
    h3_cell TEXT NOT NULL,
    h3_resolution SMALLINT NOT NULL CHECK (h3_resolution BETWEEN 0 AND 15),
    derived_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (locality_id, h3_cell),
    CONSTRAINT uq_locality_h3_cell UNIQUE (h3_cell)
);
