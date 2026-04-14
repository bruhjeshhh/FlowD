package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bruhjeshhh/flowd/internal/api"
	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/bruhjeshhh/flowd/internal/metrics"
	"github.com/bruhjeshhh/flowd/internal/worker"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func instrumentHandler(method, path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(rw, r)
		duration := time.Since(start).Seconds()
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, strconv.Itoa(rw.statusCode)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func updateQueueMetrics(ctx context.Context, q *db.Queries) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			counts, err := q.CountJobsByStatus(ctx)
			if err != nil {
				continue
			}
			for _, c := range counts {
				if c.Status.Valid {
					metrics.JobsInQueue.WithLabelValues(c.Status.String).Set(float64(c.Count))
				}
			}
		}
	}
}

func getShutdownTimeout() time.Duration {
	if s := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 15 * time.Second
}

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	dburl := os.Getenv("DB_URL")
	if dburl == "" {
		logger.Error("DB_URL is required")
		os.Exit(1)
	}

	dbz, err := sql.Open("postgres", dburl)
	if err != nil {
		logger.Error("sql open", "err", err)
		os.Exit(1)
	}
	defer dbz.Close()

	dbz.SetMaxOpenConns(25)
	dbz.SetMaxIdleConns(5)
	dbz.SetConnMaxLifetime(5 * time.Minute)

	ctx := context.Background()
	if err := dbz.PingContext(ctx); err != nil {
		logger.Error("database ping", "err", err)
		os.Exit(1)
	}

	dbQueries := db.New(dbz)
	handler := api.NewHandler(dbQueries, dbz)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", instrumentHandler("POST", "/jobs", handler.InsertJob))
	mux.HandleFunc("POST /jobs/batch", instrumentHandler("POST", "/jobs/batch", handler.BatchInsertJobs))
	mux.HandleFunc("GET /jobs", instrumentHandler("GET", "/jobs", handler.ListJobs))
	mux.HandleFunc("GET /jobs/{id}", instrumentHandler("GET", "/jobs/{id}", handler.GetJob))
	mux.HandleFunc("DELETE /jobs/{id}", instrumentHandler("DELETE", "/jobs/{id}", handler.CancelJob))
	mux.HandleFunc("POST /jobs/{id}/replay", instrumentHandler("POST", "/jobs/{id}/replay", handler.ReplayJob))
	mux.HandleFunc("GET /health", instrumentHandler("GET", "/health", handler.Health))
	mux.Handle("/metrics", promhttp.Handler())

	rateLimit, rateWindow := api.GetRateLimitConfig()
	handlerWithMiddleware := api.RateLimitMiddleware(
		api.RequestIDMiddleware(api.CORSMiddleware(mux)),
		rateLimit,
		rateWindow,
	)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           handlerWithMiddleware,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	workerCount := 4
	if s := os.Getenv("WORKER_COUNT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			workerCount = n
		}
	}

	workerCtx, workerStop := context.WithCancel(context.Background())
	defer workerStop()

	for i := 1; i <= workerCount; i++ {
		workerCfg := &worker.APIConfig{DB: dbQueries, WorkerID: i, Log: logger}
		go workerCfg.WorkerFunc(workerCtx)
	}
	rescuerCfg := &worker.APIConfig{DB: dbQueries, WorkerID: 0, Log: logger}
	go rescuerCfg.RescuerFunc(workerCtx)

	go updateQueueMetrics(workerCtx, dbQueries)

	go func() {
		logger.Info("http listening", "addr", srv.Addr, "workers", workerCount)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logger.Info("shutting down")

	shutdownTimeout := getShutdownTimeout()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	workerStop()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown", "err", err)
	}
	logger.Info("bye")
}
