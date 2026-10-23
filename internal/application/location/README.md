# Location Application Services

## Purpose

Coordinates driver location updates, H3 indexing, and location recovery or reconciliation workflows.

## Public Interface

`DriverLocationService`, `H3RecoveryService`, `H3ReconciliationWorker`, and location validation errors.

## Dependencies

`internal/config`, `internal/domain`, `internal/infrastructure/geo`, `internal/infrastructure/redis`, `internal/ports`, OpenTelemetry, and UUID.

## Invariants & Gotchas

Coordinates and sequence values are validated before state updates. H3 recovery restores Redis from authoritative driver state and remains bounded by the configured worker interval.

## Human-Approval Required

YES — location and H3 index behavior requires review.

## Related ADRs

ADR-0008, ADR-0009, ADR-0010, ADR-0019, ADR-0021.

## Known Issues / Tech Debt

Recovery is timer-based; distributed reconciliation locking is intentionally not part of the current architecture.
