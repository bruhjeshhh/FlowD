package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/google/uuid"
)

type incoming struct {
	Payload        json.RawMessage `json:"payload"`
	ScheduledAt    time.Time       `json:"scheduled_at"`
	IdempotencyKey string          `json:"idempotency_key"`
}

type payloadData struct {
	Type string `json:"type"`
	Data struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	} `json:"data"`
}

func validatePayload(pls incoming) bool {
	if len(pls.Payload) == 0 || string(pls.Payload) == "null" {
		return false
	}

	var pld payloadData
	if err := json.Unmarshal(pls.Payload, &pld); err != nil {
		return false
	}

	if pld.Type == "" {
		return false
	}
	if !pls.ScheduledAt.IsZero() {
		_, err := pls.ScheduledAt.MarshalText()
		if err != nil {
			return false
		}
	}
	return true
}

func (c *apiConfig) insertjob(w http.ResponseWriter, r *http.Request) {
	decode := json.NewDecoder(r.Body)
	pld := incoming{}
	decodingerror := decode.Decode(&pld)
	if decodingerror != nil {
		respondWithError(w, 400, "not a Json(probably)")
		return
	}
	validity := validatePayload(pld)

	if validity == false {
		respondWithError(w, 400, "invalid format")
		return
	}

	ctx := context.Background()
	params := db.InsertJobParams{
		ID:             uuid.New(),
		Payload:        pld.Payload,
		Status:         sql.NullString{String: "pending", Valid: true},
		RetryCount:     0,
		MaxRetries:     3,
		IdempotencyKey: pld.IdempotencyKey,
		ScheduledAt:    sql.NullTime{Time: pld.ScheduledAt, Valid: !pld.ScheduledAt.IsZero()},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := c.db.InsertJob(ctx, params)
	if err != nil {
		respondWithError(w, 500, "Failed to insert job")
		return
	}

	respondWithJson(w, 201, "Job created successfully")
}
