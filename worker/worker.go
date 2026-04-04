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

		log.Printf("Processing job with ID: %s", id)
		if err := c.DB.UpdateJobStatus(context.Background(), id); err != nil {
			log.Printf("worker error updating job status: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
}

