package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCancelJobInvalidUUID(t *testing.T) {
	mux := http.NewServeMux()
	h := &Handler{}
	mux.HandleFunc("DELETE /jobs/{id}", h.CancelJob)

	req := httptest.NewRequest(http.MethodDelete, "/jobs/not-a-valid-uuid", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
