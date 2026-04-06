package main

import (
	"context"
	"database/sql"
	"log"
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

	dburl := os.Getenv("DB_URL")
	if dburl == "" {
		log.Fatal("DB_URL is required")
	}

	dbz, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatal(err)
	}
	defer dbz.Close()

	dbz.SetMaxOpenConns(25)
	dbz.SetMaxIdleConns(5)
	dbz.SetConnMaxLifetime(5 * time.Minute)

	ctx := context.Background()
	if err := dbz.PingContext(ctx); err != nil {
		log.Fatalf("database ping: %v", err)
	}

	dbQueries := db.New(dbz)
	cfg := apiConfig{db: dbQueries, dbConn: dbz}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", cfg.insertjob)
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
		workerCfg := &worker.APIConfig{DB: dbQueries, WorkerID: i}
		go workerCfg.WorkerFunc(workerCtx)
	}
	rescuerCfg := &worker.APIConfig{DB: dbQueries, WorkerID: 0}
	go rescuerCfg.RescuerFunc(workerCtx)

	go func() {
		log.Printf("listening on %s (%d workers + rescuer)", srv.Addr, workerCount)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Printf("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	workerStop()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	log.Printf("bye")
}
