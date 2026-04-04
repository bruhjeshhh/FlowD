package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type incoming struct {
	Payload struct {
		Type string `json:"type"`
		Data struct {
			To      string `json:"to"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		} `json:"data"`
	} `json:"payload"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func validatePayload(pls incoming) bool {
	if pls.Payload.Type == "" {
		return false
	}
	return true
}

func (c *apiConfig) insertjob(w http.ResponseWriter, r *http.Request) {
	decode := json.NewDecoder(r.Body)
	pld := incoming{}
	decodingerror := decode.Decode(&pld)
	if decodingerror != nil {
		respondWithError(w, 400, "not a Json(probably)")
	}
	validity := validatePayload(pld)

	if validity == false {
		respondWithError(w, 400, "invalid format")
		return
	}

}
