package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithError(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	respondWithError(w, http.StatusTeapot, "short and stout")

	if w.Code != http.StatusTeapot {
		t.Fatalf("code %d", w.Code)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "short and stout" {
		t.Fatalf("error message: %q", body.Error)
	}
}

func TestRespondWithJson(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	payload := map[string]any{"id": "123", "name": "test"}
	respondWithJson(w, http.StatusOK, payload)

	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: got %q, want %q", ct, "application/json")
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["id"] != "123" {
		t.Fatalf("id: got %v, want %q", got["id"], "123")
	}
	if got["name"] != "test" {
		t.Fatalf("name: got %v, want %q", got["name"], "test")
	}
}
