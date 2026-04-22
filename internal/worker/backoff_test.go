package worker

import (
	"testing"
	"time"
)

func TestExponentialBackoffCalculation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		retryCount int
		expected  time.Duration
	}{
		{0, 5 * time.Second},
		{1, 10 * time.Second},
		{2, 20 * time.Second},
		{3, 40 * time.Second},
		{4, 80 * time.Second},
		{5, 160 * time.Second},
		{10, 5120 * time.Second},
		{20, 5242880 * time.Second},
		{29, 5368709120 * time.Second},
		{30, 5368709120 * time.Second},
	}

	for _, tt := range tests {
		actual := calculateBackoff(tt.retryCount)
		if actual != tt.expected {
			t.Errorf("retry %d: expected %v, got %v", tt.retryCount, tt.expected, actual)
		}
	}
}

func calculateBackoff(retryCount int) time.Duration {
	if retryCount >= 30 {
		retryCount = 30
	}
	delay := 5 * (1 << retryCount)
	return time.Duration(delay) * time.Second
}

func TestExponentialBackoffCapped(t *testing.T) {
	t.Parallel()

	delay29 := calculateBackoff(29)
	delay30 := calculateBackoff(30)
	delay31 := calculateBackoff(31)

	if delay29 != delay30 {
		t.Errorf("retry 29: expected %v, got %v", delay29, delay30)
	}
	if delay30 != delay31 {
		t.Errorf("retry 31: expected %v, got %v", delay30, delay31)
	}
}

func TestBackoffGrowth(t *testing.T) {
	t.Parallel()

	prev := time.Duration(0)
	for i := 0; i < 10; i++ {
		curr := calculateBackoff(i)
		if curr <= prev {
			t.Errorf("retry %d: expected larger than %v, got %v", i, prev, curr)
		}
		prev = curr
	}
}