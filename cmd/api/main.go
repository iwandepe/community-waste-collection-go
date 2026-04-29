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
	"github.com/iwandp/community-waste-collection-go/internal/repository"
	"github.com/iwandp/community-waste-collection-go/internal/service"
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

	// services
	householdSvc := service.NewHouseholdService(householdRepo)

	// handlers
	householdHandler := handler.NewHouseholdHandler(householdSvc)

	r := gin.Default()

	api := r.Group("/api")
	{
		hh := api.Group("/households")
		hh.POST("", householdHandler.Create)
		hh.GET("", householdHandler.List)
		hh.GET("/:id", householdHandler.GetByID)
		hh.DELETE("/:id", householdHandler.Delete)
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
