package worker

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/bruhjeshhh/flowd/metrics"

	db "github.com/bruhjeshhh/flowd/internal/database"
)

type APIConfig struct {
	DB       *db.Queries
	WorkerID int
	Log      *slog.Logger
}

func (c *APIConfig) logger() *slog.Logger {
	if c.Log != nil {
		return c.Log
	}
	return slog.Default()
}

func (c *APIConfig) WorkerFunc(ctx context.Context) {
	log := c.logger().With("component", "worker", "worker_id", c.WorkerID)

	metrics.WorkersActive.Inc()
	defer metrics.WorkersActive.Dec()

	for {
		select {
		case <-ctx.Done():
			log.Info("worker stopping")
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
			log.Error("get job failed", "err", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		jlog := log.With("job_id", response.ID.String(), "job_type", response.Type)
		jlog.Info("job claimed")

		start := time.Now()
		ok := handlejobs(c.logger(), response.Type, response.Payload)
		duration := time.Since(start).Seconds()

		if ok {
			if err := c.DB.UpdateJobStatusSuccess(ctx, response.ID); err != nil {
				jlog.Error("mark success failed", "err", err)
				continue
			}
			metrics.JobsProcessed.WithLabelValues(response.Type, "success").Inc()
			metrics.JobDuration.WithLabelValues(response.Type).Observe(duration)
			jlog.Info("job succeeded")
		} else {
			if err := c.DB.UpdateJobStatusNotSuccess(ctx, response.ID); err != nil {
				jlog.Error("mark failure/retry failed", "err", err)
				continue
			}
			metrics.JobsProcessed.WithLabelValues(response.Type, "failed").Inc()
			metrics.JobDuration.WithLabelValues(response.Type).Observe(duration)
			jlog.Info("job handler failed, updated for retry or terminal fail",
				"retry_count", response.RetryCount,
				"max_retries", response.MaxRetries)
		}
	}
}

func (c *APIConfig) RescuerFunc(ctx context.Context) {
	log := c.logger().With("component", "rescuer")

	for {
		select {
		case <-ctx.Done():
			log.Info("rescuer stopping")
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
			log.Error("get stuck job failed", "err", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
			continue
		}

		log.Info("resetting stuck job", "job_id", id.String())
		if err := c.DB.ResetStuckJob(ctx, id); err != nil {
			log.Error("reset stuck job failed", "job_id", id.String(), "err", err)
		}
	}
}
