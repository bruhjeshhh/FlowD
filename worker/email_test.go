package worker

import (
	"io"
	"log/slog"
	"testing"
)

func TestHandleEmails_Validation(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	_, err := handleemails(log, []byte(`{"to":"a@b.c"}`))
	if err == nil {
		t.Fatal("expected error for missing body/html")
	}

	_, err = handleemails(log, []byte(`{"body":"hi"}`))
	if err == nil {
		t.Fatal("expected error for missing 'to'")
	}

	_, err = handleemails(log, []byte(`{`))
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}
