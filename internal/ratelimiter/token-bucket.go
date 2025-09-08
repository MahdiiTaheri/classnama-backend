package ratelimiter

import (
	"sync"
	"time"
)

type tokenBucket struct {
	sync.Mutex
	tokens     float64
	lastRefill time.Time
}

type TokenBucketRateLimiter struct {
	clients sync.Map // map[ip]*tokenBucket
	rate    float64  // tokens per second
	burst   int      // bucket capacity
	window  time.Duration
}

func NewTokenBucketLimiter(reqsPerWindow int, window time.Duration) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		rate:   float64(reqsPerWindow) / window.Seconds(),
		burst:  reqsPerWindow,
		window: window,
	}
}

func (rl *TokenBucketRateLimiter) getBucket(ip string) *tokenBucket {
	val, ok := rl.clients.Load(ip)
	if ok {
		return val.(*tokenBucket)
	}
	tb := &tokenBucket{tokens: float64(rl.burst), lastRefill: time.Now()}
	actual, _ := rl.clients.LoadOrStore(ip, tb)
	return actual.(*tokenBucket)
}

func (rl *TokenBucketRateLimiter) Allow(ip string) (bool, time.Duration) {
	tb := rl.getBucket(ip)
	tb.Lock()
	defer tb.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * rl.rate
	if tb.tokens > float64(rl.burst) {
		tb.tokens = float64(rl.burst)
	}
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens -= 1
		return true, 0
	}

	wait := time.Duration((1 - tb.tokens) / rl.rate * float64(time.Second))
	return false, wait
}

// Cleanup: scan occasionally, but not blocking Allow
func (rl *TokenBucketRateLimiter) StartCleanup() {
	ticker := time.NewTicker(rl.window)
	go func() {
		for now := range ticker.C {
			rl.clients.Range(func(key, value any) bool {
				tb := value.(*tokenBucket)
				tb.Lock()
				expired := now.Sub(tb.lastRefill) > rl.window*2
				tb.Unlock()
				if expired {
					rl.clients.Delete(key)
				}
				return true
			})
		}
	}()
}
