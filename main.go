package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	db "github.com/bruhjeshhh/flowd/internal/database"
	"github.com/joho/godotenv"
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

	// ptr.HandleFunc("POST /jobs", cfg.register)

	log.Printf("we ballin")
	log.Fatal(srv.ListenAndServe())

}
