package api

import (
	"database/sql"
	"net/http"
	"strconv"

	db "github.com/bruhjeshhh/flowd/internal/database"
)

const (
	dlqListDefaultLimit = 50
	dlqListMaxLimit     = 200
)

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		respondWithError(w, http.StatusBadRequest, "query parameter status is required (use status=failed for dead-letter queue)")
		return
	}
	if status != "failed" {
		respondWithError(w, http.StatusBadRequest, "only status=failed is supported (dead-letter queue)")
		return
	}

	limit := dlqListDefaultLimit
	if ls := r.URL.Query().Get("limit"); ls != "" {
		n, err := strconv.Atoi(ls)
		if err != nil || n < 1 {
			respondWithError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		if n > dlqListMaxLimit {
			n = dlqListMaxLimit
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

	jobs, err := h.db.ListJobsByStatus(r.Context(), db.ListJobsByStatusParams{
		Status: sql.NullString{String: status, Valid: true},
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	counts, err := h.db.CountJobsByStatus(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	var total int64
	for _, c := range counts {
		if c.Status.Valid && c.Status.String == status {
			total = c.Count
		}
	}

	out := make([]jobOut, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, jobToOut(j))
	}

	respondWithJson(w, http.StatusOK, map[string]any{
		"jobs":   out,
		"limit":  limit,
		"offset": offset,
		"total":  total,
	})
}
