package main

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/google/uuid"
)

func TestJobToOut(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	scheduledAt := now.Add(time.Hour)
	nextRunAt := now.Add(30 * time.Minute)

	in := db.Job{
		ID:             id,
		Payload:        json.RawMessage(`{"type":"email","data":{"to":"a@b.c"}}`),
		Status:         sql.NullString{String: "pending", Valid: true},
		Type:           "email",
		RetryCount:     0,
		MaxRetries:     3,
		IdempotencyKey: "test-key-123",
		ScheduledAt:    sql.NullTime{Time: scheduledAt, Valid: true},
		CreatedAt:      now,
		UpdatedAt:      now,
		NextRunAt:      sql.NullTime{Time: nextRunAt, Valid: true},
	}

	out := jobToOut(in)

	if out.ID != id.String() {
		t.Fatalf("ID: got %q, want %q", out.ID, id.String())
	}
	if string(out.Payload) != string(in.Payload) {
		t.Fatalf("Payload mismatch")
	}
	if out.Type != "email" {
		t.Fatalf("Type: got %q, want %q", out.Type, "email")
	}
	if out.RetryCount != 0 {
		t.Fatalf("RetryCount: got %d, want 0", out.RetryCount)
	}
	if out.MaxRetries != 3 {
		t.Fatalf("MaxRetries: got %d, want 3", out.MaxRetries)
	}
	if out.IdempotencyKey != "test-key-123" {
		t.Fatalf("IdempotencyKey: got %q, want %q", out.IdempotencyKey, "test-key-123")
	}
	if out.Status == nil || *out.Status != "pending" {
		t.Fatalf("Status: got %v, want %q", out.Status, "pending")
	}
	if out.ScheduledAt == nil || !out.ScheduledAt.Equal(scheduledAt) {
		t.Fatalf("ScheduledAt: got %v, want %v", out.ScheduledAt, scheduledAt)
	}
	if out.NextRunAt == nil || !out.NextRunAt.Equal(nextRunAt) {
		t.Fatalf("NextRunAt: got %v, want %v", out.NextRunAt, nextRunAt)
	}
	if !out.CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt: got %v, want %v", out.CreatedAt, now)
	}
	if !out.UpdatedAt.Equal(now) {
		t.Fatalf("UpdatedAt: got %v, want %v", out.UpdatedAt, now)
	}
}

func TestJobToOut_NullFields(t *testing.T) {
	t.Parallel()
	in := db.Job{
		ID:             uuid.New(),
		Payload:        json.RawMessage(`{}`),
		Status:         sql.NullString{Valid: false},
		Type:           "sms",
		RetryCount:     1,
		MaxRetries:     5,
		IdempotencyKey: "key",
		ScheduledAt:    sql.NullTime{Valid: false},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		NextRunAt:      sql.NullTime{Valid: false},
	}

	out := jobToOut(in)

	if out.Status != nil {
		t.Fatalf("Status: got %v, want nil", out.Status)
	}
	if out.ScheduledAt != nil {
		t.Fatalf("ScheduledAt: got %v, want nil", out.ScheduledAt)
	}
	if out.NextRunAt != nil {
		t.Fatalf("NextRunAt: got %v, want nil", out.NextRunAt)
	}
}
