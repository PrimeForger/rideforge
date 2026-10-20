# Surge Policy Domain (`internal/domain/surgepolicy`)

## Purpose
Defines validated, locality-scoped surge multiplier policies and bounded imbalance-ratio tiers.

## Public Interface
`Policy`, `RatioTier`, `Status`, `ErrInvalidPolicy`, `ErrOverlappingPolicies`, `Policy.Validate`, `Policy.ValidAt`, `Policy.MultiplierFor`, and `ValidateNoOverlappingActivePolicies`.

## Dependencies
`errors`, `math`, `time`, and `github.com/google/uuid`.

## Invariants & Gotchas
An ACTIVE policy must contain at least one tier. Thresholds are strictly increasing, multipliers are nondecreasing, and each multiplier must remain within the policy maximum. Effective periods for ACTIVE policies of the same locality cannot overlap; adjacent historical, current, and future periods are valid. Phase B's HOT classification threshold (default 2 requests) is separate from the surge policy minimum demand (default 3 requests), which independently suppresses surge to 1.00x.
Policy multipliers are bounded from 1.00× through 1.75×. Higher caps are explicit policy data. Tiers are strictly ordered; minimum demand suppresses surge to 1.00×. The package has no I/O or framework dependencies.

## Human-Approval Required
YES. Policy caps and multiplier calculation affect customer pricing.

## Related ADRs
ADR-0007 (PostgreSQL as Primary Database), ADR-0021 (Failure and Degradation), ADR-0024 (Configuration and Environment Strategy).

## Known Issues / Tech Debt
No management API is included in Phase C; policy state is loaded from PostgreSQL and updated through the existing operational database workflow.
