package api

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[ip]
	var valid []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[ip] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[ip] = valid
	return true
}

func RateLimitMiddleware(next http.Handler, limit int, window time.Duration) http.Handler {
	limiter := newRateLimiter(limit, window)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		if !limiter.allow(ip) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetRateLimitConfig() (int, time.Duration) {
	limit := 100
	if s := osGetenv("RATE_LIMIT_PER_WINDOW"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}

	window := 60 * time.Second
	if s := osGetenv("RATE_LIMIT_WINDOW_SECONDS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			window = time.Duration(n) * time.Second
		}
	}

	return limit, window
}

func osGetenv(key string) string {
	v, _ := os.LookupEnv(key)
	return v
}
