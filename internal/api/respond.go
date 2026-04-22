package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/bruhjeshhh/flowd/internal/config"
	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/bruhjeshhh/flowd/internal/metrics"
	"github.com/bruhjeshhh/flowd/pkg/pqutil"
)

type Handler struct {
	db     *db.Queries
	dbConn *sql.DB
}

func NewHandler(db *db.Queries, dbConn *sql.DB) *Handler {
	return &Handler{db: db, dbConn: dbConn}
}

func (h *Handler) JobsEnqueuedInc(jobType string) {
	metrics.JobsEnqueued.WithLabelValues(jobType).Inc()
}

func (h *Handler) GetMaxRetries(jobType string) int32 {
	return config.GetMaxRetriesForJobType(jobType)
}

func isUniqueViolation(err error) bool {
	return pqutil.IsUniqueViolation(err)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorstc struct {
		Error string `json:"error"`
	}
	errormsg := errorstc{
		Error: msg,
	}

	resp, eror := json.Marshal(errormsg)
	if eror != nil {
		slog.Error("cannot unmarshal error message", "error", eror)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)

}

func respondWithJson(w http.ResponseWriter, n int, payload any) {

	resp, eror := json.Marshal(payload)
	if eror != nil {
		slog.Error("cannot marshal payload", "error", eror)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(n)
	w.Write(resp)

}
