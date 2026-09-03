package utils

import (
	"sync"
	"time"
)

// TokenBlacklist stores revoked token IDs (jti) until they expire.
// In-memory implementation suitable for single-instance deployments.
// For multi-instance deployments, replace with Redis-backed storage.
type TokenBlacklist struct {
	mu    sync.RWMutex
	items map[string]time.Time // jti -> expiration time
}

var (
	defaultBlacklist     *TokenBlacklist
	defaultBlacklistOnce sync.Once
)

// GetTokenBlacklist returns the process-wide singleton TokenBlacklist.
func GetTokenBlacklist() *TokenBlacklist {
	defaultBlacklistOnce.Do(func() {
		defaultBlacklist = &TokenBlacklist{
			items: make(map[string]time.Time),
		}
		go defaultBlacklist.cleanupLoop()
	})
	return defaultBlacklist
}

// Revoke adds a token ID to the blacklist with its expiration time.
func (b *TokenBlacklist) Revoke(jti string, expiresAt time.Time) {
	if jti == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items[jti] = expiresAt
}

// IsRevoked checks whether a token ID is in the blacklist.
func (b *TokenBlacklist) IsRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	b.mu.RLock()
	exp, ok := b.items[jti]
	b.mu.RUnlock()
	if !ok {
		return false
	}
	// Auto-cleanup if already expired
	if time.Now().After(exp) {
		b.mu.Lock()
		delete(b.items, jti)
		b.mu.Unlock()
		return false
	}
	return true
}

// cleanupLoop periodically removes expired entries to free memory.
func (b *TokenBlacklist) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		b.mu.Lock()
		now := time.Now()
		for jti, exp := range b.items {
			if now.After(exp) {
				delete(b.items, jti)
			}
		}
		b.mu.Unlock()
	}
}
