# Redis Real-Time State & Cache Adapters (`internal/infrastructure/redis`)

## Purpose
Manages transient real-time state in Redis: driver location indexes, atomic distributed locks, cell demand counters, and timeout schedulers.

## Public Interface
`Client`, `DriverLocker`, `DriverCache`, `GeoService`, `H3DriverIndex`, `H3DensityProvider`, `CellDemandStore`, `CellOccupancyStore`, `TimeoutScheduler`, `EventBus`.

## Dependencies
`github.com/redis/go-redis/v9`, `internal/ports`, `internal/infrastructure/redis/scripts`.

## Invariants & Gotchas
Redis is used for transient real-time state only (never ground-truth persistence). Demand uses a ride-to-cell pointer, per-cell ZSET, and cell registry for stale-cell recovery. All demand multi-key updates use Lua scripts; PostgreSQL remains authoritative. Driver H3 membership and the `drivers:online` registry include only available drivers; accepted rides remove supply membership, and PostgreSQL-backed H3 recovery restores eligible drivers. Occupancy Lua reads use `EVAL` in pipelines so reads work after a Redis restart clears its script cache, and Redis command errors are returned rather than converted to zero counts.

## Human-Approval Required
Yes. Modifying distributed locks or Lua script invocations directly affects driver concurrency safety.

## Related ADRs
ADR-0009 (Redis for Real-Time State), ADR-0010 (Driver Location Storage Strategy).

## Known Issues / Tech Debt
Standalone Redis instance; Redis Cluster configuration will be introduced under high load.
