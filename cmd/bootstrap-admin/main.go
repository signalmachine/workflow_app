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

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/envload"
)

const bootstrapTimeout = 30 * time.Second

func main() {
	if err := envload.LoadDefaultIfPresent(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	var (
		databaseURL     string
		orgName         string
		orgSlug         string
		userEmail       string
		userDisplayName string
		password        string
	)

	flag.StringVar(&databaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.StringVar(&orgName, "org-name", "North Harbor Works", "organization display name")
	flag.StringVar(&orgSlug, "org-slug", "north-harbor", "organization slug used for sign-in")
	flag.StringVar(&userEmail, "email", "admin@northharbor.local", "admin user email used for sign-in")
	flag.StringVar(&userDisplayName, "display-name", "North Harbor Admin", "admin user display name")
	flag.StringVar(&password, "password", "", "plain-text admin password to hash and store")
	flag.Parse()

	if databaseURL == "" {
		log.Fatal("database URL is required; set DATABASE_URL or pass -database-url")
	}
	if password == "" {
		log.Fatal("password is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), bootstrapTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	result, err := identityaccess.NewService(db).BootstrapAdmin(ctx, identityaccess.BootstrapAdminInput{
		OrgName:         orgName,
		OrgSlug:         orgSlug,
		UserEmail:       userEmail,
		UserDisplayName: userDisplayName,
		Password:        password,
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		log.Fatalf("bootstrap admin: %v", err)
	}

	fmt.Printf("org_slug=%s\n", orgSlug)
	fmt.Printf("email=%s\n", userEmail)
	fmt.Printf("org_id=%s\n", result.OrgID)
	fmt.Printf("user_id=%s\n", result.UserID)
	fmt.Printf("membership_id=%s\n", result.MembershipID)
}
