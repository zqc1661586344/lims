package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	mu       sync.Mutex
	attempts []time.Time
}

func (b *rateBucket) allow(window time.Duration, max int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
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

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
	window  time.Duration
	max     int
}

func NewRateLimiter(window time.Duration, max int) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*rateBucket),
		window:  window,
		max:     max,
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
			b.mu.Unlock()
			if recent == 0 {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !rl.getBucket(key).allow(rl.window, rl.max) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "too many login attempts, please try again later",
			})
			return
		}
		c.Next()
	}
}
