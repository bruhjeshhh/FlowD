package worker

import (
	"io"
	"log/slog"
	"testing"
)

func TestHandlejobs(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if !handlejobs(log, "email", []byte(`{}`)) {
		t.Fatal("email should succeed (stub)")
	}
	if !handlejobs(log, "sms", []byte(`{}`)) {
		t.Fatal("sms should succeed (stub)")
	}
	if !handlejobs(log, "push_notification", []byte(`{}`)) {
		t.Fatal("push_notification should succeed (stub)")
	}
	if handlejobs(log, "unknown", []byte(`{}`)) {
		t.Fatal("unknown type should fail")
	}
}
