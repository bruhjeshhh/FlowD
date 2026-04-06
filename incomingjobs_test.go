package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidatePayload(t *testing.T) {
	t.Parallel()
	goodPayload, _ := json.Marshal(map[string]any{
		"type": "email",
		"data": map[string]string{"to": "a@b.c", "subject": "hi", "body": "x"},
	})
	cases := []struct {
		name string
		in   incoming
		ok   bool
	}{
		{"empty", incoming{}, false},
		{"null payload", incoming{Payload: json.RawMessage(`null`)}, false},
		{"invalid json in payload", incoming{Payload: json.RawMessage(`{`)}, false},
		{"missing type", incoming{Payload: json.RawMessage(`{"data":{}}`)}, false},
		{"valid", incoming{Payload: goodPayload}, true},
		{"valid with schedule", incoming{Payload: goodPayload, ScheduledAt: time.Now().Add(time.Hour)}, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := validatePayload(tc.in); got != tc.ok {
				t.Fatalf("validatePayload() = %v, want %v", got, tc.ok)
			}
		})
	}
}

func TestNextRunAt(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	if got := nextRunAt(now, time.Time{}); !got.Equal(now) {
		t.Fatalf("zero scheduled: got %v want %v", got, now)
	}
	if got := nextRunAt(now, future); !got.Equal(future) {
		t.Fatalf("future scheduled: got %v want %v", got, future)
	}
	if got := nextRunAt(now, past); !got.Equal(now) {
		t.Fatalf("past scheduled: got %v want %v", got, now)
	}
}

func TestParsePayloadType(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"type":"sms","data":{}}`)
	typ, err := parsePayloadType(raw)
	if err != nil || typ != "sms" {
		t.Fatalf("parsePayloadType() = %q, %v", typ, err)
	}
	_, err = parsePayloadType(json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}
