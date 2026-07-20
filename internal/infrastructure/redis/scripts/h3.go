package scripts

import goredis "github.com/redis/go-redis/v9"

var UpdateDriverCellScript = goredis.NewScript(`
local mappingKey = KEYS[1]
local newCellSetKey = KEYS[2]

local driverID = ARGV[1]
local newCell = ARGV[2]
local ttl = tonumber(ARGV[3])
local cellPrefix = ARGV[4]

local oldCell = redis.call("GET", mappingKey)

-- New driver
if not oldCell then
    redis.call("SADD", newCellSetKey, driverID)
    redis.call("EXPIRE", newCellSetKey, ttl)

    redis.call(
        "SET",
        mappingKey,
        newCell,
        "EX",
        ttl
    )

    return {2, ""}
end

-- Same cell
if oldCell == newCell then
    redis.call("EXPIRE", newCellSetKey, ttl)
    redis.call("EXPIRE", mappingKey, ttl)

    return {0, oldCell}
end

-- Cell movement
local oldCellSetKey = cellPrefix .. oldCell

redis.call(
    "SREM",
    oldCellSetKey,
    driverID
)

redis.call(
    "SADD",
    newCellSetKey,
    driverID
)

redis.call(
    "EXPIRE",
    newCellSetKey,
    ttl
)

redis.call(
    "SET",
    mappingKey,
    newCell,
    "EX",
    ttl
)

return {1, oldCell}
`)

var RemoveDriverCellScript = goredis.NewScript(`
local mappingKey = KEYS[1]

local driverID = ARGV[1]
local cellPrefix = ARGV[2]

local currentCell = redis.call("GET", mappingKey)

if not currentCell then
    return {0, ""}
end

local cellSetKey = cellPrefix .. currentCell

redis.call(
    "SREM",
    cellSetKey,
    driverID
)

redis.call(
    "DEL",
    mappingKey
)

return {1, currentCell}
`)
