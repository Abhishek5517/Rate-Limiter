package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// token bucket
const luaScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
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

func RateLimit(Redis *redis.Client, key string, rate float64, burst int, requested int) bool {
	now := time.Now().Unix()
	result, err := Redis.Eval(context.Background(), luaScript, []string{key}, rate, burst, now, requested).Result()
	if err != nil {
		fmt.Println("Redis error:", err)
		return false
	}
	return result.(int64) == 1
}
