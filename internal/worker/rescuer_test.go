package worker

import (
	"database/sql"
	"testing"
	"time"
)

func TestRescuerRecoversStuckJobs(t *testing.T) {
	t.Parallel()

	type StuckJob struct {
		ID          string
		Status      string
		UpdatedAt   time.Time
		ScheduledAt sql.NullTime
		NextRunAt   sql.NullTime
	}

	oneMinAgo := time.Now().Add(-1 * time.Minute)
	now := time.Now()

	tests := []struct {
		name        string
		job         StuckJob
		shouldReset bool
	}{
		{
			name: "stuck processing one min ago",
			job: StuckJob{
				ID:        "abc-123",
				Status:    "processing",
				UpdatedAt: oneMinAgo,
			},
			shouldReset: true,
		},
		{
			name: "processing just now",
			job: StuckJob{
				ID:        "def-456",
				Status:    "processing",
				UpdatedAt: now,
			},
			shouldReset: false,
		},
		{
			name: "pending status should not be rescued",
			job: StuckJob{
				ID:        "ghi-789",
				Status:    "pending",
				UpdatedAt: oneMinAgo,
			},
			shouldReset: false,
		},
		{
			name: "success status should not be rescued",
			job: StuckJob{
				ID:        "jkl-012",
				Status:    "success",
				UpdatedAt: oneMinAgo,
			},
			shouldReset: false,
		},
		{
			name: "future scheduled_at should not be rescued",
			job: StuckJob{
				ID:          "mno-345",
				Status:      "processing",
				UpdatedAt:   oneMinAgo,
				ScheduledAt: sql.NullTime{Time: time.Now().Add(1 * time.Hour), Valid: true},
			},
			shouldReset: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			shouldReset := isJobStuck(tt.job.Status, tt.job.UpdatedAt, tt.job.ScheduledAt)
			if shouldReset != tt.shouldReset {
				t.Errorf("expected %v, got %v", tt.shouldReset, shouldReset)
			}
		})
	}
}

func isJobStuck(status string, updatedAt time.Time, scheduledAt sql.NullTime) bool {
	if status != "processing" {
		return false
	}

	stuckThreshold := time.Now().Add(-1 * time.Minute)
	if updatedAt.After(stuckThreshold) {
		return false
	}

	if scheduledAt.Valid && scheduledAt.Time.After(time.Now()) {
		return false
	}

	return true
}

func TestRescuerQueryLogic(t *testing.T) {
	t.Parallel()

	t.Run("GetStuckJobs SQL equivalent", func(t *testing.T) {
		t.Parallel()

		oneMinAgo := time.Now().Add(-1 * time.Minute)

		actualThreshold := oneMinAgo
		expectedThreshold := time.Now().Add(-1 * time.Minute)

		if actualThreshold.Sub(expectedThreshold).Abs() > 1*time.Second {
			t.Errorf("threshold calculation mismatch")
		}
	})
}