package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	db "github.com/bruhjeshhh/flowd/internal/database"
)

const (
	webhookListDefaultLimit = 50
	webhookListMaxLimit     = 200
)

type webhookIn struct {
	URL     string `json:"url"`
	JobType string `json:"job_type"`
	Event   string `json:"event"`
	Secret  string `json:"secret,omitempty"`
}

type webhookOut struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	JobType  string    `json:"job_type"`
	Event    string    `json:"event"`
	Secret   string    `json:"secret,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var in webhookIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if in.URL == "" {
		respondWithError(w, http.StatusBadRequest, "url is required")
		return
	}
	if in.JobType == "" {
		respondWithError(w, http.StatusBadRequest, "job_type is required")
		return
	}
	if in.Event != "job_success" && in.Event != "job_failed" {
		respondWithError(w, http.StatusBadRequest, "event must be job_success or job_failed")
		return
	}

	webhook, err := h.db.InsertWebhook(r.Context(), db.InsertWebhookParams{
		ID:        uuid.New(),
		URL:       in.URL,
		JobType:   in.JobType,
		Event:     in.Event,
		Secret:    sql.NullString{String: in.Secret, Valid: in.Secret != ""},
		CreatedAt: time.Now(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create webhook")
		return
	}

	respondWithJson(w, http.StatusCreated, map[string]any{
		"webhook": webhookToOut(webhook),
	})
}

func (h *Handler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	limit := webhookListDefaultLimit
	if ls := r.URL.Query().Get("limit"); ls != "" {
		n, err := strconv.Atoi(ls)
		if err != nil || n < 1 {
			respondWithError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		if n > webhookListMaxLimit {
			n = webhookListMaxLimit
		}
		limit = n
	}

	offset := 0
	if os := r.URL.Query().Get("offset"); os != "" {
		n, err := strconv.Atoi(os)
		if err != nil || n < 0 {
			respondWithError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = n
	}

	webhooks, err := h.db.ListWebhooks(r.Context(), db.ListWebhooksParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	count, err := h.db.CountWebhooks(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	out := make([]webhookOut, 0, len(webhooks))
	for _, w := range webhooks {
		out = append(out, webhookToOut(w))
	}

	respondWithJson(w, http.StatusOK, map[string]any{
		"webhooks": out,
		"limit":   limit,
		"offset":  offset,
		"total":   count,
	})
}

func (h *Handler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "webhook id is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid webhook id")
		return
	}

	if err := h.db.DeleteWebhook(r.Context(), id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to delete webhook")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func webhookToOut(w db.Webhook) webhookOut {
	out := webhookOut{
		ID:        w.ID.String(),
		URL:       w.URL,
		JobType:  w.JobType,
		Event:    w.Event,
		CreatedAt: w.CreatedAt,
	}
	if w.Secret.Valid {
		out.Secret = "***"
	}
	return out
}