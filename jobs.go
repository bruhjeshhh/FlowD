package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type incoming struct {
	Payload        json.RawMessage `json:"payload"`
	ScheduledAt    time.Time       `json:"scheduled_at"`
	IdempotencyKey string          `json:"idempotency_key"`
}

type payloadData struct {
	Type string `json:"type"`
	Data struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	} `json:"data"`
}

func validatePayload(pls incoming) bool {
	if len(pls.Payload) == 0 || string(pls.Payload) == "null" {
		return false
	}

	var pld payloadData
	if err := json.Unmarshal(pls.Payload, &pld); err != nil {
		return false
	}

	if pld.Type == "" {
		return false
	}
	if !pls.ScheduledAt.IsZero() {
		_, err := pls.ScheduledAt.MarshalText()
		if err != nil {
			return false
		}
	}
	return true
}

func (c *apiConfig) insertjob(w http.ResponseWriter, r *http.Request) {
	decode := json.NewDecoder(r.Body)
	pld := incoming{}
	decodingerror := decode.Decode(&pld)
	if decodingerror != nil {
		respondWithError(w, 400, "not a Json(probably)")
		return
	}
	validity := validatePayload(pld)

	if validity == false {
		respondWithError(w, 400, "invalid format")
		return
	}

}
