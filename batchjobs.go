package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/bruhjeshhh/flowd/metrics"
	"github.com/google/uuid"
)

type batchJobRequest struct {
	Jobs []incoming `json:"jobs"`
}

type batchJobResult struct {
	Success    bool   `json:"success"`
	Job        jobOut `json:"job,omitempty"`
	Error      string `json:"error,omitempty"`
	Idempotent bool   `json:"idempotent_replay,omitempty"`
}

func (c *apiConfig) batchInsertJobs(w http.ResponseWriter, r *http.Request) {
	decode := json.NewDecoder(r.Body)
	var req batchJobRequest
	if err := decode.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if len(req.Jobs) == 0 {
		respondWithError(w, http.StatusBadRequest, "jobs array is required and cannot be empty")
		return
	}

	if len(req.Jobs) > 100 {
		respondWithError(w, http.StatusBadRequest, "maximum 100 jobs per batch")
		return
	}

	ctx := r.Context()
	now := time.Now().UTC()

	type jobResult struct {
		index  int
		result batchJobResult
	}

	results := make([]batchJobResult, len(req.Jobs))
	typeCh := make(chan jobResult, len(req.Jobs))

	for i, pld := range req.Jobs {
		go func(idx int, p incoming) {
			if !validatePayload(p) {
				typeCh <- jobResult{idx, batchJobResult{
					Success: false,
					Error:   "invalid format",
				}}
				return
			}

			jobType, err := parsePayloadType(p.Payload)
			if err != nil {
				typeCh <- jobResult{idx, batchJobResult{
					Success: false,
					Error:   "invalid payload",
				}}
				return
			}

			idempotencyKey := p.IdempotencyKey
			if idempotencyKey == "" {
				idempotencyKey = uuid.New().String()
			}

			tx, err := c.dbConn.BeginTx(ctx, nil)
			if err != nil {
				typeCh <- jobResult{idx, batchJobResult{
					Success: false,
					Error:   "database error",
				}}
				return
			}
			defer func() { _ = tx.Rollback() }()

			qtx := c.db.WithTx(tx)

			existing, err := qtx.GetJobByIdempotencyKey(ctx, idempotencyKey)
			if err == nil {
				_ = tx.Commit()
				typeCh <- jobResult{idx, batchJobResult{
					Success:    true,
					Job:        jobToOut(existing),
					Idempotent: true,
				}}
				return
			}
			if !isUniqueViolation(err) && err != sql.ErrNoRows {
				typeCh <- jobResult{idx, batchJobResult{
					Success: false,
					Error:   "database error",
				}}
				return
			}

			params := db.InsertJobParams{
				ID:             uuid.New(),
				Payload:        p.Payload,
				Status:         sql.NullString{String: "pending", Valid: true},
				Type:           jobType,
				RetryCount:     0,
				MaxRetries:     GetMaxRetriesForJobType(jobType),
				IdempotencyKey: idempotencyKey,
				ScheduledAt:    sql.NullTime{Time: p.ScheduledAt, Valid: !p.ScheduledAt.IsZero()},
				CreatedAt:      now,
				UpdatedAt:      now,
				NextRunAt:      sql.NullTime{Time: nextRunAt(now, p.ScheduledAt), Valid: true},
			}

			job, err := qtx.InsertJob(ctx, params)
			if err != nil {
				if isUniqueViolation(err) {
					existing, err2 := qtx.GetJobByIdempotencyKey(ctx, idempotencyKey)
					if err2 != nil {
						typeCh <- jobResult{idx, batchJobResult{
							Success: false,
							Error:   "database error",
						}}
						return
					}
					_ = tx.Commit()
					typeCh <- jobResult{idx, batchJobResult{
						Success:    true,
						Job:        jobToOut(existing),
						Idempotent: true,
					}}
					return
				}
				typeCh <- jobResult{idx, batchJobResult{
					Success: false,
					Error:   "failed to create job",
				}}
				return
			}

			if err := tx.Commit(); err != nil {
				typeCh <- jobResult{idx, batchJobResult{
					Success: false,
					Error:   "database error",
				}}
				return
			}

			metrics.JobsEnqueued.WithLabelValues(jobType).Inc()

			typeCh <- jobResult{idx, batchJobResult{
				Success: true,
				Job:     jobToOut(job),
			}}
		}(i, pld)
	}

	for i := 0; i < len(req.Jobs); i++ {
		res := <-typeCh
		results[res.index] = res.result
	}

	respondWithJson(w, http.StatusCreated, map[string]interface{}{
		"results": results,
		"total":   len(req.Jobs),
	})
}
