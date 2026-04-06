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

func (c *APIConfig) WorkerFunc(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		response, err := c.DB.GetJobByScheduledAt(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
				}
				continue
			}
			log.Printf("worker[%d] error getting job: %v", c.WorkerID, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		ok := handlejobs(response.Type, response.Payload)
		if ok {
			if err := c.DB.UpdateJobStatusSuccess(ctx, response.ID); err != nil {
				log.Printf("worker[%d] error updating job status: %v", c.WorkerID, err)
			}
		} else {
			if err := c.DB.UpdateJobStatusNotSuccess(ctx, response.ID); err != nil {
				log.Printf("worker[%d] error updating job status: %v", c.WorkerID, err)
			}
		}
	}
}

func (c *APIConfig) RescuerFunc(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		id, err := c.DB.GetStuckJobs(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				select {
				case <-ctx.Done():
					return
				case <-time.After(1 * time.Minute):
				}
				continue
			}
			log.Printf("rescuer error getting stuck job: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
			continue
		}

		if err := c.DB.ResetStuckJob(ctx, id); err != nil {
			log.Printf("rescuer error resetting stuck job: %v", err)
		}
	}
}
