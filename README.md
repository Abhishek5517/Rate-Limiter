# Rate Limiter

A Redis-backed rate limiting library for Go, implementing two algorithms: **Token Bucket** and **Sliding Window**. Both algorithms run as atomic Lua scripts in Redis, making them safe for distributed, multi-instance deployments.

## Algorithms

### Token Bucket
Tokens accumulate at a fixed `rate` up to a `burst` capacity. Each request consumes tokens. If there aren't enough tokens, the request is denied. This allows short bursts of traffic while enforcing a long-term average rate.

### Sliding Window
Tracks individual request timestamps in a Redis sorted set. A request is allowed only if the number of requests within the last `windowSize` duration is below `maxRequests`. This enforces a strict per-window limit with no boundary spikes.

## Installation

```bash
go get github.com/Abhishek5517/Rate-Limiter
```

Requires a running Redis instance.

## Usage

### Token Bucket

```go
import (
    ratelimiter "github.com/Abhishek5517/Rate-Limiter"
    "github.com/go-redis/redis/v8"
)

redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// 10 tokens/sec refill rate, burst capacity of 50
rl := ratelimiter.NewTokenBucket("user:123", 10, 50)

if rl.TokenBucket(redisClient, 1) {
    // request allowed
} else {
    // rate limited
}
```

### Sliding Window

```go
import (
    ratelimiter "github.com/Abhishek5517/Rate-Limiter"
    "github.com/go-redis/redis/v8"
    "time"
)

redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// max 100 requests per minute
rl := ratelimiter.NewSlidingWindow("user:123", time.Minute, 100)

if rl.SlidingWindow(redisClient) {
    // request allowed
} else {
    // rate limited
}
```

## API

### Constructors

```go
func NewTokenBucket(key string, rate float64, burst int) *RateLimiter
```
- `key` — unique identifier for the rate-limited entity (e.g. user ID, IP address)
- `rate` — token refill rate in tokens per second
- `burst` — maximum bucket capacity and initial token count

```go
func NewSlidingWindow(key string, windowSize time.Duration, maxRequests int) *RateLimiter
```
- `key` — unique identifier for the rate-limited entity
- `windowSize` — duration of the sliding window
- `maxRequests` — maximum number of requests allowed within the window

### Methods

```go
func (rl *RateLimiter) TokenBucket(Redis *redis.Client, requested int) bool
```
Returns `true` if `requested` tokens are available and the request is allowed.

```go
func (rl *RateLimiter) SlidingWindow(Redis *redis.Client) bool
```
Returns `true` if the request count within the current window is below `maxRequests`.

Both methods return `false` on Redis errors.

## Algorithm Comparison

| | Token Bucket | Sliding Window |
|---|---|---|
| Redis structure | Hash (`HMSET`) | Sorted set (`ZADD`) |
| Memory per key | O(1) | O(requests in window) |
| Burst support | Yes — explicit burst capacity | No — limit is strictly per window |
| Use case | Smooth bursty traffic | Strict per-window enforcement |
| Distributed safe | Yes — atomic Lua script | Yes — atomic Lua script |

## Running Tests

Tests use [miniredis](https://github.com/alicebob/miniredis) — no running Redis instance required.

```bash
go test ./...
```

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/go-redis/redis/v8` | Redis client |
| `github.com/google/uuid` | Unique request IDs for sliding window |
| `github.com/alicebob/miniredis/v2` | In-memory Redis for tests |
