package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	mu       sync.Mutex
	tokens   float64
	lastSeen time.Time
	rate     float64 // tokens per second
	capacity float64
}

func (b *bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastSeen = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

var globalBuckets sync.Map

// RateLimit returns a token-bucket middleware keyed per API key (Authorization header).
// Falls back to RemoteAddr for unauthenticated endpoints.
func RateLimit(requests int, per time.Duration) func(http.Handler) http.Handler {
	rate := float64(requests) / per.Seconds()
	capacity := float64(requests)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Authorization")
			if key == "" {
				key = r.RemoteAddr
			}
			// Namespace by rate so different limits get independent buckets.
			bucketKey := fmt.Sprintf("%s|%d|%.4f", key, requests, rate)

			v, _ := globalBuckets.LoadOrStore(bucketKey, &bucket{
				tokens:   capacity,
				lastSeen: time.Now(),
				rate:     rate,
				capacity: capacity,
			})
			b := v.(*bucket)

			if !b.allow() {
				w.Header().Set("Retry-After", "60")
				writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
