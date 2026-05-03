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

	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/bruhjeshhh/flowd/worker"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db     *db.Queries
	dbConn *sql.DB
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
	cfg := apiConfig{db: dbQueries, dbConn: dbz}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", cfg.insertjob)
	mux.HandleFunc("GET /jobs", cfg.listJobs)
	mux.HandleFunc("GET /jobs/{id}", cfg.getJob)
	mux.HandleFunc("GET /health", cfg.health)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	workerStop()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown", "err", err)
	}
	logger.Info("bye")
}
