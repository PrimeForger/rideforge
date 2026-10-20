# Locality Foundation Status

This file tracks the locality foundation independently of Stage 7 status.

## Implemented

- Generic locality domain model with stable UUID identity and draft/active/inactive lifecycle.
- Locality creation accepts an authoritative boundary definition and does not accept customer-provided locality IDs.
- Root PostgreSQL migrations `000003_locality_foundation` and `000004_surge_policy` create the locality tables, derived H3 mappings, locality-scoped policies, and policy tiers.
- PostGIS `MultiPolygon` geometry is authoritative. H3 coverage uses the existing operational resolution, and boundary plus derived mapping changes run in one PostgreSQL transaction.
- Each H3 cell is uniquely owned in the derived mapping table. Resolution returns one active locality, unresolved when there is no match, and an explicit error for ambiguous matches. Activation requires an operational-resolution H3 mapping.
- Per-locality policy defaults are 1.00x minimum multiplier, 1.50x maximum multiplier, and 3 minimum requests. A 1.75x maximum is accepted only when explicit policy data sets it.
- ACTIVE policies require at least one valid tier. Thresholds are strictly increasing; multipliers are nondecreasing and bounded by the policy maximum.
- Migration `000005_surge_policy_nonoverlap` uses a partial GiST exclusion constraint and half-open time ranges to reject overlapping ACTIVE policies for the same locality. Historical and adjacent current/future periods are allowed; the runtime cache also rejects duplicate effective policies.
- Failed policy refreshes preserve the last known-good snapshot and do not advance its successful-load time. The snapshot is rejected after configured maximum staleness; each application instance refreshes independently from PostgreSQL.
- Surge calculation uses Phase B occupancy and reports bounded-cardinality fallback outcomes. Unavailable or invalid inputs return the 1.00x base multiplier.
- Phase B HOT classification defaults to 2 requests. Phase C's independent policy minimum demand defaults to 3 requests, so a HOT cell can still use the base multiplier.

## Deferred

- Admin authentication/authorization and HTTP admin handlers/routes.
- Quote/fare integration. No fare quote operation currently exists; adding one or a pricing API is outside Phase C scope.
- Redis policy caching and policy events.

## Verification

On 2026-10-20, focused Phase C and Phase B occupancy tests passed. `go test -count=1 -ldflags=-linkmode=internal ./...`, `go vet ./...`, and `go test -cover -ldflags=-linkmode=internal ./...` passed using `.tmp-go-build` as the Go cache. `internal/domain/surgepolicy` coverage was 82.9%.

The race test was attempted but Windows linking failed because `runtime/race` requires external linking. At the time of the Phase C review, migration `005_surge_policy_nonoverlap` had not been applied because PostgreSQL/PostGIS was unavailable.

Later on 2026-10-20, root migrations `000001`–`000005` passed fresh apply, complete rollback, and reapply on a disposable PostGIS 15 database. The resulting application schema matches the reference snapshot.
