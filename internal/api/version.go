package api

import (
	"net/http"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
}

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, VersionResponse{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
	})
}