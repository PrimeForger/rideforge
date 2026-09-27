# Search Profiles (`internal/application/dispatch/discovery/profile`)

## Purpose
Provides context-aware search profiles for varying operational environments (Urban, Rural, Airport, Stand).

## Public Interface
`SearchProfile` struct, `ExpansionProfile` struct, `ExpansionConfig` struct, `ExpansionProfileProvider` interface, `DefaultExpansionProfileProvider`, `ProfileBuilder`, `DefaultExpansionPolicy`.

## Dependencies
`internal/config`.

## Invariants & Gotchas
Stand and airport profiles prioritize drivers in designated queues before expanding to nearby street drivers.

## Human-Approval Required
Yes. Search profiles dictate operational and geographic matching rules.

## Related ADRs
ADR-0015 (Smart Dispatch & Stand Dispatch).

## Known Issues / Tech Debt
Profiles are statically configured; dynamic database-driven profiles planned.
