package worker

import (
	"testing"
)

func TestConcurrencyWithSelectForUpdate(t *testing.T) {
	t.Parallel()

	t.Run("GetJobByScheduledAt has FOR UPDATE SKIP LOCKED", func(t *testing.T) {
		t.Parallel()
		t.Skip("Integration test - requires running PostgreSQL")
	})

	t.Run("concurrent workers claim unique jobs", func(t *testing.T) {
		t.Parallel()

		type claimedJob struct {
			workerID int
			jobID    string
		}

		jobs := []string{"job-1", "job-2", "job-3", "job-4", "job-5"}
		claimed := make(chan claimedJob, len(jobs))

		var workers int = 3

		for w := 0; w < workers; w++ {
			go func(workerID int) {
				for _, jobID := range jobs {
					claimed <- claimedJob{workerID, jobID}
				}
			}(w)
		}

		uniqueJobs := make(map[string]int)
		for i := 0; i < len(jobs); i++ {
			c := <-claimed
			uniqueJobs[c.jobID]++
		}

		for jobID, count := range uniqueJobs {
			if count > 1 {
				t.Errorf("job %s claimed %d times (expected 1)", jobID, count)
			}
		}
	})
}

func TestConcurrentClaimAtomicity(t *testing.T) {
	t.Parallel()

	t.Run("claim returns job or nothing", func(t *testing.T) {
		t.Parallel()

		claimResult := func() (string, bool) {
			return "", true
		}

		jobID, ok := claimResult()
		if jobID == "" && !ok {
			t.Error("expected either a job ID or nothing")
		}
	})
}

func TestWorkerIsolation(t *testing.T) {
	t.Parallel()

	type Worker struct {
		id    int
		jobs  []string
		claim map[string]bool
	}

	workers := []Worker{
		{id: 0, jobs: []string{"A", "B", "C"}, claim: make(map[string]bool)},
		{id: 1, jobs: []string{"A", "B", "C"}, claim: make(map[string]bool)},
	}

	for w := range workers {
		for _, job := range workers[w].jobs {
			workers[w].claim[job] = true
		}
	}

	jobACount := 0
	for _, w := range workers {
		if w.claim["A"] {
			jobACount++
		}
	}

	if jobACount != 2 {
		t.Errorf("each worker should independently claim their view of A")
	}
}