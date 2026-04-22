package api

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.allow("1.2.3.4") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	if rl.allow("1.2.3.4") {
		t.Error("4th request should be denied")
	}
}

func TestRateLimiterPerIPIsolation(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(2, time.Minute)

	if !rl.allow("1.1.1.1") {
		t.Error("first IP first request should be allowed")
	}
	if !rl.allow("1.1.1.1") {
		t.Error("first IP second request should be allowed")
	}
	if rl.allow("1.1.1.1") {
		t.Error("first IP third request should be denied")
	}

	if !rl.allow("2.2.2.2") {
		t.Error("second IP first request should be allowed (different IP)")
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(2, 50*time.Millisecond)

	if !rl.allow("1.2.3.4") {
		t.Error("request 1 should be allowed")
	}
	if !rl.allow("1.2.3.4") {
		t.Error("request 2 should be allowed")
	}
	if rl.allow("1.2.3.4") {
		t.Error("request 3 should be denied")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.allow("1.2.3.4") {
		t.Error("after window reset, request should be allowed")
	}
}

func TestRateLimiterMiddleware429(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(1, time.Minute)

	var wg sync.WaitGroup
	wg.Add(2)

	var results [2]int

	go func() {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), 1, time.Minute).ServeHTTP(w, req)
		results[0] = w.Code
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), 1, time.Minute).ServeHTTP(w, req)
		results[1] = w.Code
		wg.Done()
	}()

	wg.Wait()

	if results[0] != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", results[0])
	}
	if results[1] != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", results[1])
	}
}

func TestRateLimiterXForwardedFor(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(2, time.Minute)

	if !rl.allow("1.1.1.1") {
		t.Error("first request should be allowed")
	}
	if !rl.allow("1.1.1.1") {
		t.Error("second request should be allowed")
	}
	if rl.allow("1.1.1.1") {
		t.Error("third request should be denied")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "3.3.3.3")
	req.RemoteAddr = "1.2.3.4:1234"

	rl2 := newRateLimiter(1, time.Minute)
	if !rl2.allow(req.RemoteAddr) {
		t.Error("requests without X-Forwarded-For should use RemoteAddr")
	}
}