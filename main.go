package ratelimiter

import (
	"context"
	"fmt"
	"time"

	tokenbucket "github.com/Abhishek5517/Rate-Limiter/TokenBucket"
	"github.com/go-redis/redis/v8"
)

func RateLimit(Redis *redis.Client, key string, rate float64, burst int, requested int) bool {
	now := time.Now().UnixNano()
	result, err := Redis.Eval(context.Background(), tokenbucket.LuaScript, []string{key}, rate, burst, now, requested).Result()
	if err != nil {
		fmt.Println("Redis error:", err)
		return false
	}
	allowed, ok := result.(int64)
	if !ok {
		return false
	}
	return allowed == 1
}
