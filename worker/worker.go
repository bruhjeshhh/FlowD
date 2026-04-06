package worker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	db "github.com/bruhjeshhh/flowd/internal/database"
)

type APIConfig struct {
	DB       *db.Queries
	WorkerID int
}

func (c *APIConfig) WorkerFunc() {
	for {
		response, err := c.DB.GetJobByScheduledAt(context.Background())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				time.Sleep(2 * time.Second)
				continue
			}
			log.Printf("worker[%d] error getting job: %v", c.WorkerID, err)
			time.Sleep(5 * time.Second)
			continue
		}

		doneornot := handlejobs(response.Type, response.Payload)
		if doneornot {
			if err := c.DB.UpdateJobStatusSuccess(context.Background(), response.ID); err != nil {
				log.Printf("worker[%d] error updating job status: %v", c.WorkerID, err)
			}

		} else {
			if err := c.DB.UpdateJobStatusNotSuccess(context.Background(), response.ID); err != nil {
				log.Printf("worker[%d] error updating job status: %v", c.WorkerID, err)
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
