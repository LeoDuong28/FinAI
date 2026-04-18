package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter implements per-IP token bucket rate limiting.
type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*ipLimiter
	rate     rate.Limit
	burst    int
	done     chan struct{}
}

// NewRateLimiter creates a rate limiter with the given requests/second and burst.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		rate:     rate.Limit(rps),
		burst:    burst,
		done:     make(chan struct{}),
	}

	go rl.cleanup()

	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if l, exists := rl.limiters[ip]; exists {
		l.lastSeen = time.Now()
		return l.limiter
	}

	l := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, l := range rl.limiters {
				if time.Since(l.lastSeen) > 10*time.Minute {
					delete(rl.limiters, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// Handler returns the rate limiting middleware.
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip port from RemoteAddr (e.g. "192.168.1.1:54321" → "192.168.1.1")
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.burst))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"RATE_LIMITED","message":"Too many requests. Please slow down."}}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
