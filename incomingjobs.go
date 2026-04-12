package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/bruhjeshhh/flowd/metrics"
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
		if _, err := pls.ScheduledAt.MarshalText(); err != nil {
			return false
		}
	}
	return true
}

func parsePayloadType(payload json.RawMessage) (string, error) {
	var pld payloadData
	if err := json.Unmarshal(payload, &pld); err != nil {
		return "", err
	}
	if pld.Type == "" {
		return "", errors.New("missing type")
	}
	return pld.Type, nil
}

func nextRunAt(now time.Time, scheduledAt time.Time) time.Time {
	if !scheduledAt.IsZero() && scheduledAt.After(now) {
		return scheduledAt
	}
	return now
}

func (c *apiConfig) insertjob(w http.ResponseWriter, r *http.Request) {
	decode := json.NewDecoder(r.Body)
	pld := incoming{}
	if err := decode.Decode(&pld); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !validatePayload(pld) {
		respondWithError(w, http.StatusBadRequest, "invalid format")
		return
	}

	jobType, err := parsePayloadType(pld.Payload)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	if pld.IdempotencyKey == "" {
		pld.IdempotencyKey = uuid.New().String()
	}

	ctx := r.Context()
	now := time.Now().UTC()

	tx, err := c.dbConn.BeginTx(ctx, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	qtx := c.db.WithTx(tx)

	existing, err := qtx.GetJobByIdempotencyKey(ctx, pld.IdempotencyKey)
	if err == nil {
		if err := tx.Commit(); err != nil {
			respondWithError(w, http.StatusInternalServerError, "database error")
			return
		}
		respondWithJson(w, http.StatusOK, createJobResponse{
			Job:              jobToOut(existing),
			IdempotentReplay: true,
		})
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	params := db.InsertJobParams{
		ID:             uuid.New(),
		Payload:        pld.Payload,
		Status:         sql.NullString{String: "pending", Valid: true},
		Type:           jobType,
		RetryCount:     0,
		MaxRetries:     GetMaxRetriesForJobType(jobType),
		IdempotencyKey: pld.IdempotencyKey,
		ScheduledAt:    sql.NullTime{Time: pld.ScheduledAt, Valid: !pld.ScheduledAt.IsZero()},
		CreatedAt:      now,
		UpdatedAt:      now,
		NextRunAt:      sql.NullTime{Time: nextRunAt(now, pld.ScheduledAt), Valid: true},
	}

	job, err := qtx.InsertJob(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			existing, err2 := qtx.GetJobByIdempotencyKey(ctx, pld.IdempotencyKey)
			if err2 != nil {
				respondWithError(w, http.StatusInternalServerError, "database error")
				return
			}
			if err := tx.Commit(); err != nil {
				respondWithError(w, http.StatusInternalServerError, "database error")
				return
			}
			respondWithJson(w, http.StatusOK, createJobResponse{
				Job:              jobToOut(existing),
				IdempotentReplay: true,
			})
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	metrics.JobsEnqueued.WithLabelValues(jobType).Inc()

	respondWithJson(w, http.StatusCreated, createJobResponse{
		Job:              jobToOut(job),
		IdempotentReplay: false,
	})
}
