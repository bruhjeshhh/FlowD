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
	DB       *db.Queries
	WorkerID int
}

func (c *APIConfig) WorkerFunc() {
	for {
		id, err := c.DB.GetJobByScheduledAt(context.Background())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				time.Sleep(2 * time.Second)
				continue
			}
			log.Printf("worker[%d] error getting job: %v", c.WorkerID, err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("worker[%d] processing job with id: %v", c.WorkerID, id)
		if rand.IntN(2) == 0 {
			count, err := c.DB.IncrementRetryCount(context.Background(), id)
			if err != nil {
				log.Printf("worker[%d] error incrementing retry count for job with id: %v, error: %v", c.WorkerID, id, err)
			}
			if count >= 3 {
				log.Printf("worker[%d] job with id: %v has reached maximum retry attempts, marking as failed", c.WorkerID, id)
			} else {
				log.Printf("worker[%d] processing failed for job with id: %v, retrying...", c.WorkerID, id)
			}
			if err := c.DB.UpdateJobStatusNotSuccess(context.Background(), id); err != nil {
				log.Printf("worker[%d] error updating job status to not success for id: %v, error: %v", c.WorkerID, id, err)
			}
		} else {
			log.Printf("worker[%d] job with id: %v has been successfully completed, marking as success", c.WorkerID, id)
			if err := c.DB.UpdateJobStatusSuccess(context.Background(), id); err != nil {
				log.Printf("worker[%d] error updating job status to success for id: %v, error: %v", c.WorkerID, id, err)
			}
		}

	}
}

func (c *APIConfig) RescuerFunc() {
	for {
		id, err := c.DB.GetStuckJobs(context.Background())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				time.Sleep(1 * time.Minute)
				continue
			}
			log.Printf("rescuer error getting stuck job: %v", err)
			if err := c.DB.ResetStuckJob(context.Background(), id); err != nil {
				log.Printf("rescuer error resetting stuck job: %v", err)
			}
		}
	}
}
