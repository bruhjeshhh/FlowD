//go:build integration

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	db "github.com/bruhjeshhh/flowd/internal/database"
	_ "github.com/lib/pq"
)

func TestIntegrationIdempotentCreate(t *testing.T) {
	dburl := os.Getenv("DB_URL")
	if dburl == "" {
		t.Skip("set DB_URL for integration tests")
	}

	dbz, err := sql.Open("postgres", dburl)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dbz.Close() })

	if err := dbz.Ping(); err != nil {
		t.Fatal(err)
	}

	cfg := apiConfig{db: db.New(dbz), dbConn: dbz}

	body := map[string]any{
		"idempotency_key": "integration-test-" + t.Name(),
		"payload": map[string]any{
			"type": "email",
			"data": map[string]string{"to": "x@y.z", "subject": "s", "body": "b"},
		},
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cfg.insertjob(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("first POST: %d %s", w.Code, w.Body.String())
	}

	var first createJobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if first.IdempotentReplay {
		t.Fatal("first request should not be a replay")
	}

	req2 := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(b))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	cfg.insertjob(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second POST: %d %s", w2.Code, w2.Body.String())
	}

	var second createJobResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &second); err != nil {
		t.Fatal(err)
	}
	if !second.IdempotentReplay {
		t.Fatal("second request should be idempotent replay")
	}
	if second.Job.ID != first.Job.ID {
		t.Fatalf("job id mismatch: %q vs %q", second.Job.ID, first.Job.ID)
	}
}
