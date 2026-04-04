package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

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
	workerCfg := &worker.APIConfig{DB: dbQueries}
	go workerCfg.WorkerFunc()
	ptr.HandleFunc("POST /jobs", cfg.insertjob)

	log.Printf("we ballin")
	log.Fatal(srv.ListenAndServe())

}
