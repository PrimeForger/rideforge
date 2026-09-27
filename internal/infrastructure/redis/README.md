# Redis Real-Time State & Cache Adapters (`internal/infrastructure/redis`)

## Purpose
Manages transient real-time state in Redis: driver location indexes, atomic distributed locks, cell demand counters, and timeout schedulers.

## Public Interface
`Client`, `DriverLocker`, `DriverCache`, `GeoService`, `H3DriverIndex`, `H3DensityProvider`, `CellDemandStore`, `CellOccupancyStore`, `TimeoutScheduler`, `EventBus`.

## Dependencies
`github.com/redis/go-redis/v9`, `internal/ports`, `internal/infrastructure/redis/scripts`.

## Invariants & Gotchas
Redis is used for transient real-time state only (never ground-truth persistence). All keys must have explicit TTLs. All multi-key updates must use Lua scripts.

## Human-Approval Required
Yes. Modifying distributed locks or Lua script invocations directly affects driver concurrency safety.

## Related ADRs
ADR-0009 (Redis for Real-Time State), ADR-0010 (Driver Location Storage Strategy).

## Known Issues / Tech Debt
Standalone Redis instance; Redis Cluster configuration will be introduced under high load.
