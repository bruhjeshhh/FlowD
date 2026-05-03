package main

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
