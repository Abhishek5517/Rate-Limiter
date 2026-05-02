package ratelimiter

import (
	"context"
	"fmt"
	"time"

	slidingwindow "github.com/Abhishek5517/Rate-Limiter/SlidingWindow"
	tokenbucket "github.com/Abhishek5517/Rate-Limiter/TokenBucket"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RateLimiter struct {
	rate        float64
	burst       int
	windowSize  time.Duration
	maxRequests int
}

func NewTokenBucket(rate float64, burst int) *RateLimiter {
	return &RateLimiter{rate: rate, burst: burst}
}

func NewSlidingWindow(windowSize time.Duration, maxRequests int) *RateLimiter {
	return &RateLimiter{windowSize: windowSize, maxRequests: maxRequests}
}

func (rl *RateLimiter) TokenBucket(Redis *redis.Client, key string, requested int) bool {
	now := time.Now().UnixNano()
	result, err := Redis.Eval(context.Background(), tokenbucket.LuaScript, []string{key}, rl.rate, rl.burst, now, requested).Result()
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

func (rl *RateLimiter) SlidingWindow(Redis *redis.Client, key string) bool {
	now := time.Now().UnixNano()
	uuidWithHyphen := uuid.New()
	reqId := fmt.Sprintf("%d:%s", now, uuidWithHyphen.String())
	result, err := Redis.Eval(context.Background(), slidingwindow.LuaScript, []string{key}, now, rl.windowSize.Nanoseconds(), rl.maxRequests, reqId).Result()
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
