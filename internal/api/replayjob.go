package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/bruhjeshhh/flowd/internal/metrics"
	"github.com/google/uuid"
)

func (h *Handler) ReplayJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	ctx := r.Context()

	dlqJob, err := h.db.GetDeadLetterJobByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "dead letter job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	tx, err := h.dbConn.BeginTx(ctx, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	qtx := h.db.WithTx(tx)

	now := time.Now().UTC()
	idempotencyKey := dlqJob.IdempotencyKey.String + "-replay-" + uuid.New().String()

	newJobParams := db.InsertJobParams{
		ID:             uuid.New(),
		Payload:        dlqJob.Payload,
		Status:         sql.NullString{String: "pending", Valid: true},
		Type:           dlqJob.Type,
		RetryCount:     0,
		MaxRetries:     dlqJob.MaxRetries,
		IdempotencyKey: idempotencyKey,
		ScheduledAt:    dlqJob.ScheduledAt,
		CreatedAt:      now,
		UpdatedAt:      now,
		NextRunAt:      sql.NullTime{Time: now, Valid: true},
	}

	newJob, err := qtx.InsertJob(ctx, newJobParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create replay job")
		return
	}

	if err := qtx.DeleteDeadLetterJob(ctx, id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to remove dead letter job")
		return
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	metrics.JobsEnqueued.WithLabelValues(newJob.Type).Inc()

	respondWithJson(w, http.StatusCreated, map[string]interface{}{
		"job":    jobToOut(newJob),
		"status": "replayed",
	})
}
