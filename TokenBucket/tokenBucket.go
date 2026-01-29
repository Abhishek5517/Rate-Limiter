package tokenbucket

const LuaScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])/1000000000
local requested = tonumber(ARGV[4])

local bucket = redis.call("HMGET", key, "tokens", "timestamp")
local tokens = tonumber(bucket[1]) or burst
local last_refreshed = tonumber(bucket[2]) or now

local delta = math.max(0, now - last_refreshed)
tokens = math.min(burst, tokens + delta * rate)

if tokens < requested then
    return 0
else
    tokens = tokens - requested
    redis.call("HMSET", key, "tokens", tokens, "timestamp", now)
    redis.call("EXPIRE", key, math.ceil(burst / rate))
    return 1
end
`
