package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iwandp/community-waste-collection-go/internal/handler"
	"github.com/iwandp/community-waste-collection-go/internal/middleware"
	"github.com/iwandp/community-waste-collection-go/internal/repository"
	"github.com/iwandp/community-waste-collection-go/internal/service"
	"github.com/iwandp/community-waste-collection-go/internal/storage"
	"github.com/iwandp/community-waste-collection-go/internal/worker"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	db := mustConnectDB()
	defer db.Close()

	// repositories
	householdRepo := repository.NewHouseholdRepository(db)
	pickupRepo := repository.NewPickupRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	// storage
	s3, err := storage.NewS3Storage(
		getEnv("MINIO_ENDPOINT", "localhost:9000"),
		getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		getEnv("MINIO_SECRET_KEY", "minioadmin"),
		getEnv("MINIO_BUCKET", "payments"),
		getEnv("MINIO_USE_SSL", "false") == "true",
	)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}

	// services
	householdSvc := service.NewHouseholdService(householdRepo)
	pickupSvc := service.NewPickupService(pickupRepo, paymentRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, pickupRepo, s3)

	// handlers
	householdHandler := handler.NewHouseholdHandler(householdSvc)
	pickupHandler := handler.NewPickupHandler(pickupSvc)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	// background worker — shares lifecycle with the server
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	organicWorker := worker.NewOrganicCancelWorker(pickupRepo, time.Minute)
	go organicWorker.Run(workerCtx)

	r := gin.Default()

	api := r.Group("/api")
	{
		hh := api.Group("/households")
		hh.POST("", householdHandler.Create)
		hh.GET("", householdHandler.List)
		hh.GET("/:id", householdHandler.GetByID)
		hh.DELETE("/:id", householdHandler.Delete)

		ph := api.Group("/pickups")
		ph.POST("", middleware.PickupRateLimit(), pickupHandler.Create)
		ph.GET("", pickupHandler.List)
		ph.PUT("/:id/schedule", pickupHandler.Schedule)
		ph.PUT("/:id/complete", pickupHandler.Complete)
		ph.PUT("/:id/cancel", pickupHandler.Cancel)

		pmh := api.Group("/payments")
		pmh.POST("", paymentHandler.Create)
		pmh.GET("", paymentHandler.List)
		pmh.PUT("/:id/confirm", paymentHandler.Confirm)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := getEnv("APP_PORT", "8080")
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func mustConnectDB() *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5433"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "waste_collection"),
		getEnv("DB_SSLMODE", "disable"),
	)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
