package ratelimiter

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func newTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

// TokenBucket tests

func TestTokenBucket_AllowsWithinBurst(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	rl := NewTokenBucket("tb:user1", 10, 6)

	for i := 0; i < 5; i++ {
		if !rl.TokenBucket(client, 1) {
			t.Fatalf("request %d should have been allowed (within burst)", i+1)
		}
	}
}

func TestTokenBucket_DeniesWhenBurstExceeded(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	rl := NewTokenBucket("tb:user2", 10, 3)

	// Drain the bucket
	for i := 0; i < 3; i++ {
		rl.TokenBucket(client, 1)
	}

	if rl.TokenBucket(client, 1) {
		t.Fatal("request should have been denied after burst exhausted")
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	// rate=2 tokens/sec, burst=2
	rl := NewTokenBucket("tb:user3", 2, 2)

	// Drain fully
	rl.TokenBucket(client, 1)
	rl.TokenBucket(client, 1)

	if rl.TokenBucket(client, 1) {
		t.Fatal("should be denied immediately after draining")
	}

	// Advance miniredis time by 1 second to simulate token refill
	mr.FastForward(1 * time.Second)

	if !rl.TokenBucket(client, 1) {
		t.Fatal("should be allowed after refill time has passed")
	}
}

func TestTokenBucket_DeniesWhenRequestedExceedsBurst(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	rl := NewTokenBucket("tb:user4", 10, 3)

	// Requesting more than burst in a single call should always be denied
	if rl.TokenBucket(client, 5) {
		t.Fatal("request exceeding burst capacity should be denied")
	}
}

func TestTokenBucket_IsolatedByKey(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	rl1 := NewTokenBucket("tb:userA", 10, 1)
	rl2 := NewTokenBucket("tb:userB", 10, 1)

	rl1.TokenBucket(client, 1) // drain userA

	if !rl2.TokenBucket(client, 1) {
		t.Fatal("userB should be unaffected by userA's bucket")
	}
}

// --- SlidingWindow tests ---

func TestSlidingWindow_AllowsWithinLimit(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	rl := NewSlidingWindow("sw:user1", 10*time.Second, 5)

	for i := 0; i < 5; i++ {
		if !rl.SlidingWindow(client) {
			t.Fatalf("request %d should have been allowed (within limit)", i+1)
		}
	}
}

func TestSlidingWindow_DeniesWhenLimitExceeded(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	rl := NewSlidingWindow("sw:user2", 10*time.Second, 3)

	for i := 0; i < 3; i++ {
		rl.SlidingWindow(client)
	}

	if rl.SlidingWindow(client) {
		t.Fatal("request should be denied after limit exhausted")
	}
}

func TestSlidingWindow_AllowsAfterWindowExpires(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	rl := NewSlidingWindow("sw:user3", 1*time.Second, 2)

	rl.SlidingWindow(client)
	rl.SlidingWindow(client)

	if rl.SlidingWindow(client) {
		t.Fatal("should be denied within the window")
	}

	// Advance time past the window
	mr.FastForward(2 * time.Second)

	if !rl.SlidingWindow(client) {
		t.Fatal("should be allowed after the window has passed")
	}
}

func TestSlidingWindow_IsolatedByKey(t *testing.T) {
	client, mr := newTestRedis(t)
	defer mr.Close()

	rl1 := NewSlidingWindow("sw:userA", 10*time.Second, 1)
	rl2 := NewSlidingWindow("sw:userB", 10*time.Second, 1)

	rl1.SlidingWindow(client) // exhaust userA

	if !rl2.SlidingWindow(client) {
		t.Fatal("userB should be unaffected by userA's window")
	}
}
