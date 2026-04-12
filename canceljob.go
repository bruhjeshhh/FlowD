package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func (c *apiConfig) cancelJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := c.db.CancelJob(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "job not found or cannot be cancelled")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]interface{}{
		"job":    jobToOut(job),
		"status": "cancelled",
	})
}
