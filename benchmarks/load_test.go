package flowd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"
)

func RunLoadTest(baseURL string, concurrency, totalRequests int) {
	fmt.Printf("\n=== Load Test Configuration ===\n")
	fmt.Printf("URL: %s\n", baseURL)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Total Requests: %d\n\n", totalRequests)

	payload := map[string]any{
		"idempotency_key": "bench",
		"payload": map[string]any{
			"type": "email",
			"data": map[string]string{
				"to":      "test@example.com",
				"subject": "Test",
				"body":    "Benchmark test",
			},
		},
	}
	body, _ := json.Marshal(payload)

	var success int64
	var errors int64
	var totalDuration time.Duration
	var minLatency time.Duration = 1<<63 - 1
	var maxLatency time.Duration
	latencies := make([]time.Duration, 0, totalRequests)
	mu := &sync.Mutex{}

	var wg sync.WaitGroup
	requestsPerWorker := totalRequests / concurrency

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				reqStart := time.Now()
				resp, err := http.Post(baseURL+"/jobs", "application/json", bytes.NewReader(body))
				reqDuration := time.Since(reqStart)

				mu.Lock()
				latencies = append(latencies, reqDuration)
				totalDuration += reqDuration
				if reqDuration < minLatency {
					minLatency = reqDuration
				}
				if reqDuration > maxLatency {
					maxLatency = reqDuration
				}
				if err == nil && resp.StatusCode < 400 {
					atomic.AddInt64(&success, 1)
					resp.Body.Close()
				} else {
					atomic.AddInt64(&errors, 1)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Printf("Status: %d\n", resp.StatusCode)
						resp.Body.Close()
					}
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(start)

	fmt.Printf("=== Load Test Results ===\n")
	fmt.Printf("Total Duration: %v\n", totalTime.Round(time.Millisecond))
	fmt.Printf("Successful Requests: %d\n", success)
	fmt.Printf("Failed Requests: %d\n", errors)
	fmt.Printf("Requests/Second: %.2f\n", float64(totalRequests)/totalTime.Seconds())
	fmt.Printf("Avg Latency: %v\n", (totalDuration / time.Duration(totalRequests)).Round(time.Microsecond))
	fmt.Printf("Min Latency: %v\n", minLatency.Round(time.Microsecond))
	fmt.Printf("Max Latency: %v\n", maxLatency.Round(time.Microsecond))

	if len(latencies) > 0 {
		p50 := latencies[len(latencies)/2]
		p95 := latencies[int(float64(len(latencies))*0.95)]
		p99 := latencies[int(float64(len(latencies))*0.99)]
		fmt.Printf("P50 Latency: %v\n", p50.Round(time.Microsecond))
		fmt.Printf("P95 Latency: %v\n", p95.Round(time.Microsecond))
		fmt.Printf("P99 Latency: %v\n", p99.Round(time.Microsecond))
	}

	fmt.Printf("\nThroughput: %.2f req/sec | Avg Response: %v\n",
		float64(totalRequests)/totalTime.Seconds(),
		(totalDuration / time.Duration(totalRequests)).Round(time.Microsecond))
}

func ExampleRunLoadTest() {
	RunLoadTest("http://localhost:8080", 10, 1000)
}

type MockFlowD struct {
	server *httptest.Server
}

func NewMockFlowD() *MockFlowD {
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Microsecond * 100)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"job": map[string]any{"id": "test"}})
	})
	return &MockFlowD{
		server: httptest.NewServer(mux),
	}
}

func (m *MockFlowD) Close() {
	m.server.Close()
}
