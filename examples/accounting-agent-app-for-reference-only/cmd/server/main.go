package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	webAdapter "accounting-agent/internal/adapters/web"
	"accounting-agent/internal/ai"
	"accounting-agent/internal/app"
	"accounting-agent/internal/core"
	"accounting-agent/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	ruleEngine := core.NewRuleEngine(pool)
	orderService := core.NewOrderService(pool, ruleEngine)
	inventoryService := core.NewInventoryService(pool, ruleEngine)
	reportingService := core.NewReportingService(pool)
	userService := core.NewUserService(pool)
	vendorService := core.NewVendorService(pool)
	purchaseOrderService := core.NewPurchaseOrderService(pool)

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("Warning: OPENAI_API_KEY is not set")
	}
	agent := ai.NewAgent(apiKey)

	svc := app.NewAppService(pool, ledger, docService, orderService, inventoryService, reportingService, userService, vendorService, purchaseOrderService, agent)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET is not set — refusing to start. Set a strong random value in .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}
	if port == "" {
		port = "8080"
	}

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	handler := webAdapter.NewHandler(ctx, svc, allowedOrigins, jwtSecret)

	server := &http.Server{
		Addr:              ":" + port, // binds on 0.0.0.0:<port>
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("server starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("server shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown with error: %v", err)
	}
}
