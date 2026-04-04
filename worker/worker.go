package worker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"math/rand/v2"
	"time"

	db "github.com/bruhjeshhh/flowd/internal/database"
)

type APIConfig struct {
	DB *db.Queries
}

func (c *APIConfig) WorkerFunc() {
	for {
		id, err := c.DB.GetJobByScheduledAt(context.Background())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				time.Sleep(2 * time.Second)
				continue
			}
			log.Printf("worker error getting job: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("processing job with id: %v", id)
		if rand.IntN(2) == 0 {
			count, err := c.DB.IncrementRetryCount(context.Background(), id)
			if err != nil {
				log.Printf("worker error incrementing retry count for job with id: %v, error: %v", id, err)
			}
			if count >= 3 {
				log.Printf("job with id: %v has reached maximum retry attempts, marking as failed", id)
			} else {
				log.Printf("processing failed for job with id: %v, retrying...", id)
			}
			c.DB.UpdateJobStatusNotSuccess(context.Background(), id)
		} else {
			log.Printf("job with id: %v has been successfully completed, marking as success", id)
			c.DB.UpdateJobStatusSuccess(context.Background(), id)
		}

	}
}
