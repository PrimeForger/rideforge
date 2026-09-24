package scripts

import goredis "github.com/redis/go-redis/v9"

var AddActiveDemandScript = goredis.NewScript(`
local mappingKey = KEYS[1]
local cellZSetKey = KEYS[2]

local rideID = ARGV[1]
local cellID = ARGV[2]
local score = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

redis.call("SET", mappingKey, cellID, "EX", ttl)
redis.call("ZADD", cellZSetKey, score, rideID)
redis.call("EXPIRE", cellZSetKey, ttl)

return 1
`)

var RemoveActiveDemandScript = goredis.NewScript(`
local mappingKey = KEYS[1]

local rideID = ARGV[1]
local cellPrefix = ARGV[2]
local cellSuffix = ARGV[3]

local cellID = redis.call("GET", mappingKey)

if cellID then
    local cellZSetKey = cellPrefix .. cellID .. cellSuffix
    redis.call("ZREM", cellZSetKey, rideID)
    redis.call("DEL", mappingKey)
    return cellID
end

return ""
`)

var GetAndPruneCellDemandScript = goredis.NewScript(`
local cellZSetKey = KEYS[1]
local cutoffScore = tonumber(ARGV[1])

if cutoffScore and cutoffScore > 0 then
    redis.call("ZREMRANGEBYSCORE", cellZSetKey, "-inf", cutoffScore)
end

return redis.call("ZCARD", cellZSetKey)
`)
