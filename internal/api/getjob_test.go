package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lib/pq"
)

func TestGetJob_InvalidUUID(t *testing.T) {
	t.Parallel()
	h := &Handler{}
	req := httptest.NewRequest(http.MethodGet, "/jobs/not-a-uuid", nil)
	w := httptest.NewRecorder()
	h.GetJob(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestNextRunAt_PastScheduled(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)

	if got := nextRunAt(now, past); !got.Equal(now) {
		t.Fatalf("nextRunAt(): got %v, want %v", got, now)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	t.Parallel()
	err := &pq.Error{Code: "23505"}
	if !isUniqueViolation(err) {
		t.Fatal("expected true")
	}
	if isUniqueViolation(errors.New("other")) {
		t.Fatal("expected false")
	}
}
