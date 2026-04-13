package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.db.GetJobByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "job not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	respondWithJson(w, http.StatusOK, jobToOut(job))
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.dbConn.PingContext(ctx); err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"status": "ok"})
}
