package worker

import (
	"io"
	"log/slog"
	"testing"
)

func TestHandlejobs(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if res := handlejobs(log, "email", []byte(`{}`)); !res.Success {
		t.Fatal("email should succeed (stub)")
	}
	if res := handlejobs(log, "sms", []byte(`{}`)); !res.Success {
		t.Fatal("sms should succeed (stub)")
	}
	if res := handlejobs(log, "push_notification", []byte(`{}`)); !res.Success {
		t.Fatal("push_notification should succeed (stub)")
	}
	if res := handlejobs(log, "unknown", []byte(`{}`)); res.Success {
		t.Fatal("unknown type should fail")
	}
}
