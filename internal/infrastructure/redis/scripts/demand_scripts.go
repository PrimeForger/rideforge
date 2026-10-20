package scripts

import goredis "github.com/redis/go-redis/v9"

var AddActiveDemandScript = goredis.NewScript(`
local mappingKey = KEYS[1]
local cellZSetKey = KEYS[2]
local registryKey = KEYS[3]

local rideID = ARGV[1]
local cellID = ARGV[2]
local score = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])
local cellPrefix = ARGV[5]
local cellSuffix = ARGV[6]

local oldCell = redis.call("GET", mappingKey)
if oldCell and oldCell ~= cellID then
    local oldCellKey = cellPrefix .. oldCell .. cellSuffix
    redis.call("ZREM", oldCellKey, rideID)
    if redis.call("ZCARD", oldCellKey) == 0 then
        redis.call("DEL", oldCellKey)
        redis.call("SREM", registryKey, oldCell)
    end
end

redis.call("SET", mappingKey, cellID, "EX", ttl)
redis.call("ZADD", cellZSetKey, score, rideID)
redis.call("EXPIRE", cellZSetKey, ttl)
redis.call("SADD", registryKey, cellID)

return 1
`)

var RemoveActiveDemandScript = goredis.NewScript(`
local mappingKey = KEYS[1]
local registryKey = KEYS[2]

local rideID = ARGV[1]
local cellPrefix = ARGV[2]
local cellSuffix = ARGV[3]
local fallbackCellID = ARGV[4]

local cellID = redis.call("GET", mappingKey)
if not cellID and fallbackCellID ~= "" then
    cellID = fallbackCellID
end

if cellID then
    local cellZSetKey = cellPrefix .. cellID .. cellSuffix
    redis.call("ZREM", cellZSetKey, rideID)
    if redis.call("GET", mappingKey) == cellID then
        redis.call("DEL", mappingKey)
    end
    if redis.call("ZCARD", cellZSetKey) == 0 then
        redis.call("DEL", cellZSetKey)
        redis.call("SREM", registryKey, cellID)
    end
    return cellID
end

return ""
`)

var ReconcileCellDemandScript = goredis.NewScript(`
local cellZSetKey = KEYS[1]
local registryKey = KEYS[2]
local cellID = ARGV[1]
local cellPrefix = ARGV[2]
local cellSuffix = ARGV[3]
local rideDemandPrefix = ARGV[4]
local ttl = tonumber(ARGV[5])
local cutoff = tonumber(ARGV[6])
local expectedCount = tonumber(ARGV[7])
local expected = {}
local argIndex = 8

for i = 1, expectedCount do
    expected[ARGV[argIndex]] = ARGV[argIndex + 1]
    argIndex = argIndex + 2
end

local members = redis.call("ZRANGE", cellZSetKey, 0, -1)
for _, rideID in ipairs(members) do
    if not expected[rideID] then
        redis.call("ZREM", cellZSetKey, rideID)
        local mappingKey = rideDemandPrefix .. rideID
        if redis.call("GET", mappingKey) == cellID then
            redis.call("DEL", mappingKey)
        end
    end
end

for rideID, score in pairs(expected) do
    local mappingKey = rideDemandPrefix .. rideID
    local oldCell = redis.call("GET", mappingKey)
    if oldCell and oldCell ~= cellID then
        local oldCellKey = cellPrefix .. oldCell .. cellSuffix
        redis.call("ZREM", oldCellKey, rideID)
        if redis.call("ZCARD", oldCellKey) == 0 then
            redis.call("DEL", oldCellKey)
            redis.call("SREM", registryKey, oldCell)
        end
    end
    redis.call("ZADD", cellZSetKey, score, rideID)
    redis.call("SET", mappingKey, cellID, "EX", ttl)
end

redis.call("ZREMRANGEBYSCORE", cellZSetKey, "-inf", cutoff)
if redis.call("ZCARD", cellZSetKey) == 0 then
    redis.call("DEL", cellZSetKey)
    redis.call("SREM", registryKey, cellID)
else
    redis.call("EXPIRE", cellZSetKey, ttl)
    redis.call("SADD", registryKey, cellID)
end

return redis.call("ZCARD", cellZSetKey)
`)

var GetAndPruneCellDemandScript = goredis.NewScript(`
local cellZSetKey = KEYS[1]
local cutoffScore = tonumber(ARGV[1])

if cutoffScore and cutoffScore > 0 then
    redis.call("ZREMRANGEBYSCORE", cellZSetKey, "-inf", cutoffScore)
end

return redis.call("ZCARD", cellZSetKey)
`)
