package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/bruhjeshhh/flowd/worker"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db *db.Queries
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dburl := os.Getenv("DB_URL")

	dbz, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatal(err)
	}
	defer dbz.Close()
	dbQueries := db.New(dbz)
	var cfg apiConfig

	cfg.db = dbQueries
	ptr := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: ptr,
	}

	workerCount := 4
	if s := os.Getenv("WORKER_COUNT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			workerCount = n
		}
	}

	for i := 1; i <= workerCount; i++ {
		workerCfg := &worker.APIConfig{DB: dbQueries, WorkerID: i}
		go workerCfg.WorkerFunc()
	}
	rescuerCfg := &worker.APIConfig{DB: dbQueries, WorkerID: 0}
	go rescuerCfg.RescuerFunc()
	ptr.HandleFunc("POST /jobs", cfg.insertjob)

	log.Printf("we ballin with %d workers and a rescuer", workerCount)
	log.Fatal(srv.ListenAndServe())

}
