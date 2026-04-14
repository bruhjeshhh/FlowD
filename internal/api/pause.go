package api

import (
	"net/http"
	"sync/atomic"
)

var workerPaused atomic.Bool

func (h *Handler) PauseWorkers(w http.ResponseWriter, r *http.Request) {
	workerPaused.Store(true)
	respondWithJson(w, http.StatusOK, map[string]string{"status": "workers paused"})
}

func (h *Handler) ResumeWorkers(w http.ResponseWriter, r *http.Request) {
	workerPaused.Store(false)
	respondWithJson(w, http.StatusOK, map[string]string{"status": "workers resumed"})
}

func (h *Handler) WorkerStatus(w http.ResponseWriter, r *http.Request) {
	status := "running"
	if workerPaused.Load() {
		status = "paused"
	}
	respondWithJson(w, http.StatusOK, map[string]string{"status": status})
}

func AreWorkersPaused() bool {
	return workerPaused.Load()
}
