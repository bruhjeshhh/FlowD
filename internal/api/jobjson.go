package api

import (
	"encoding/json"
	"time"

	db "github.com/bruhjeshhh/flowd/internal/database"
)

type jobOut struct {
	ID             string          `json:"id"`
	Payload        json.RawMessage `json:"payload"`
	Status         *string         `json:"status,omitempty"`
	Type           string          `json:"type"`
	RetryCount     int32           `json:"retry_count"`
	MaxRetries     int32           `json:"max_retries"`
	IdempotencyKey string          `json:"idempotency_key"`
	ScheduledAt    *time.Time      `json:"scheduled_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	NextRunAt      *time.Time      `json:"next_run_at,omitempty"`
}

type createJobResponse struct {
	Job              jobOut `json:"job"`
	IdempotentReplay bool   `json:"idempotent_replay"`
}

func jobToOut(j db.Job) jobOut {
	out := jobOut{
		ID:             j.ID.String(),
		Payload:        j.Payload,
		Type:           j.Type,
		RetryCount:     j.RetryCount,
		MaxRetries:     j.MaxRetries,
		IdempotencyKey: j.IdempotencyKey,
		CreatedAt:      j.CreatedAt,
		UpdatedAt:      j.UpdatedAt,
	}
	if j.Status.Valid {
		s := j.Status.String
		out.Status = &s
	}
	if j.ScheduledAt.Valid {
		t := j.ScheduledAt.Time
		out.ScheduledAt = &t
	}
	if j.NextRunAt.Valid {
		t := j.NextRunAt.Time
		out.NextRunAt = &t
	}
	return out
}
