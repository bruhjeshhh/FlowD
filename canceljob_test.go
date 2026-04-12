package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCancelJobInvalidUUID(t *testing.T) {
	mux := http.NewServeMux()
	cfg := apiConfig{}
	mux.HandleFunc("DELETE /jobs/{id}", cfg.cancelJob)

	req := httptest.NewRequest(http.MethodDelete, "/jobs/not-a-valid-uuid", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
