package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"workflow_app/internal/identityaccess"
)

const setPasswordTimeout = 10 * time.Second

func main() {
	var (
		databaseURL string
		userID      string
		password    string
	)

	flag.StringVar(&databaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.StringVar(&userID, "user-id", "", "identityaccess.users.id to update")
	flag.StringVar(&password, "password", "", "plain-text password to hash and store")
	flag.Parse()

	if databaseURL == "" {
		log.Fatal("database URL is required; set DATABASE_URL or pass -database-url")
	}
	if userID == "" {
		log.Fatal("user-id is required")
	}
	if password == "" {
		log.Fatal("password is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), setPasswordTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	if err := identityaccess.NewService(db).SetUserPassword(ctx, identityaccess.SetUserPasswordInput{
		UserID:    userID,
		Password:  password,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		log.Fatalf("set password: %v", err)
	}

	log.Printf("password updated for user %s", userID)
}
