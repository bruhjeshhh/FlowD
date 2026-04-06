package worker

import "testing"

func TestHandlejobs(t *testing.T) {
	t.Parallel()
	if !handlejobs("email", []byte(`{}`)) {
		t.Fatal("email should succeed (stub)")
	}
	if !handlejobs("sms", []byte(`{}`)) {
		t.Fatal("sms should succeed (stub)")
	}
	if !handlejobs("push_notification", []byte(`{}`)) {
		t.Fatal("push_notification should succeed (stub)")
	}
	if handlejobs("unknown", []byte(`{}`)) {
		t.Fatal("unknown type should fail")
	}
}
