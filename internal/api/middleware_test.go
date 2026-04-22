package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDMiddlewareGeneratesID(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestRequestIDMiddlewareUsesExistingID(t *testing.T) {
	t.Parallel()

	existingID := "existing-request-id-12345"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", existingID)
	w := httptest.NewRecorder()

	RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != existingID {
		t.Errorf("expected %q, got %q", existingID, w.Header().Get("X-Request-ID"))
	}
}

func TestCORSMiddlewareSetsHeaders(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})).ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Access-Control-Allow-Origin should be *")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Access-Control-Allow-Methods should be set")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Access-Control-Allow-Headers should be set")
	}
}

func TestCORSPreflightReturns204(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()

	CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS")
	})).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestMiddlewareChain(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler := RequestIDMiddleware(CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID should be set in chained middleware")
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS headers should be set in chained middleware")
	}
}

func TestRequestIDInContext(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "test-id-123")
	w := httptest.NewRecorder()

	var foundID string
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	foundID = r.Context().Value(RequestIDKey).(string)
	}))

	handler.ServeHTTP(w, req)

	if foundID != "test-id-123" {
		t.Errorf("expected test-id-123, got %q", foundID)
	}
}

func TestRequestIDFormat(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})).ServeHTTP(w, req)

		id := w.Header().Get("X-Request-ID")
		if len(id) != 16 {
			t.Errorf("expected 16 char hex ID, got %q (%d)", id, len(id))
		}
		if !strings.HasPrefix(id, "test-id-123") && i == 0 {
			continue
		}
	}
}