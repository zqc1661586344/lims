package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	mu          sync.Mutex
	attempts    []time.Time
	lockedUntil time.Time
}

func (b *rateBucket) allow(window time.Duration, max int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.lockedUntil.IsZero() && time.Now().Before(b.lockedUntil) {
		return false
	}
	if !b.lockedUntil.IsZero() {
		b.lockedUntil = time.Time{}
	}

	now := time.Now()
	cutoff := now.Add(-window)
	i := 0
	for i < len(b.attempts) && b.attempts[i].Before(cutoff) {
		i++
	}
	b.attempts = b.attempts[i:]
	if len(b.attempts) >= max {
		return false
	}
	b.attempts = append(b.attempts, now)
	return true
}

func (b *rateBucket) record(window time.Duration, maxFailures int, lockDuration time.Duration) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)
	i := 0
	for i < len(b.attempts) && b.attempts[i].Before(cutoff) {
		i++
	}
	b.attempts = b.attempts[i:]
	b.attempts = append(b.attempts, now)

	if len(b.attempts) >= maxFailures {
		b.lockedUntil = now.Add(lockDuration)
		return true
	}
	return false
}

func (b *rateBucket) isLocked() (bool, time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.lockedUntil.IsZero() {
		return false, 0
	}
	remaining := time.Until(b.lockedUntil)
	if remaining <= 0 {
		b.lockedUntil = time.Time{}
		return false, 0
	}
	return true, remaining
}

type RateLimiter struct {
	mu           sync.Mutex
	buckets      map[string]*rateBucket
	window       time.Duration
	max          int
	maxFailures  int
	lockDuration time.Duration
}

func NewRateLimiter(window time.Duration, max int) *RateLimiter {
	rl := &RateLimiter{
		buckets:      make(map[string]*rateBucket),
		window:       window,
		max:          max,
		maxFailures:  5,
		lockDuration: 15 * time.Minute,
	}
	go rl.gc()
	return rl
}

func (rl *RateLimiter) getBucket(key string) *rateBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.buckets[key]
	if !ok {
		b = &rateBucket{}
		rl.buckets[key] = b
	}
	return b
}

func (rl *RateLimiter) gc() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for k, b := range rl.buckets {
			b.mu.Lock()
			recent := 0
			for _, t := range b.attempts {
				if t.After(cutoff) {
					recent++
				}
			}
			locked := !b.lockedUntil.IsZero() && time.Now().Before(b.lockedUntil)
			b.mu.Unlock()
			if recent == 0 && !locked {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) AllowKey(key string) bool {
	if locked, _ := rl.IsLocked(key); locked {
		return false
	}
	return rl.getBucket(key).allow(rl.window, rl.max)
}

// RecordFail records one failure for the given key. When the failure count
// within the window reaches maxFailures the key is locked for lockDuration.
// Returns true if the account is now locked.
func (rl *RateLimiter) RecordFail(key string) bool {
	return rl.getBucket(key).record(rl.window, rl.maxFailures, rl.lockDuration)
}

// IsLocked reports whether the given key is currently locked and for how long.
func (rl *RateLimiter) IsLocked(key string) (bool, time.Duration) {
	return rl.getBucket(key).isLocked()
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		if locked, remaining := rl.IsLocked(key); locked {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("尝试次数过多，请在 %d 分钟后重试", int(remaining.Minutes())+1),
			})
			return
		}

		if !rl.AllowKey(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "登录请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}
