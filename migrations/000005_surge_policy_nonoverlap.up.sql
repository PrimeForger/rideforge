CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE surge_policies
    ADD CONSTRAINT ex_surge_policies_active_effective_period
    EXCLUDE USING gist (
        locality_id WITH =,
        tstzrange(effective_from, effective_until, '[)') WITH &&
    )
    WHERE (status = 'ACTIVE');
