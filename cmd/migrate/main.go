package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"workflow_app/internal/platform/envload"
	"workflow_app/internal/platform/migrations"
)

func main() {
	if err := envload.LoadDefaultIfPresent(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	var databaseURL string
	flag.StringVar(&databaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.Parse()

	if databaseURL == "" {
		log.Fatal("database URL is required; set DATABASE_URL or pass -database-url")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	applied, err := migrations.Up(ctx, db)
	if err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	fmt.Printf("applied %d migration(s)\n", applied)
}
