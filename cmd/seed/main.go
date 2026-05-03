package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()
	db := mustConnectDB()
	defer db.Close()

	log.Println("seeding...")

	if err := seed(db); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Println("done.")
}

func seed(db *sqlx.DB) error {
	// Households
	households := []struct{ id, owner, address string }{
		{uuid.NewString(), "Budi Santoso", "Jl. Mawar No. 12, Jakarta Selatan"},
		{uuid.NewString(), "Siti Rahayu", "Jl. Melati No. 5, Bandung"},
		{uuid.NewString(), "Ahmad Fauzi", "Jl. Kenanga No. 8, Surabaya"},
		{uuid.NewString(), "Dewi Lestari", "Jl. Anggrek No. 3, Yogyakarta"},
		{uuid.NewString(), "Riko Pratama", "Jl. Dahlia No. 21, Medan"},
	}

	for _, h := range households {
		_, err := db.Exec(
			`INSERT INTO households (id, owner_name, address, created_at, updated_at)
			 VALUES ($1, $2, $3, NOW(), NOW()) ON CONFLICT DO NOTHING`,
			h.id, h.owner, h.address,
		)
		if err != nil {
			return fmt.Errorf("insert household: %w", err)
		}
	}
	log.Printf("  inserted %d households", len(households))

	// Waste Pickups
	now := time.Now()
	pickupDate := now.Add(2 * 24 * time.Hour) // 2 days from now
	safetyTrue := true

	type pickup struct {
		id, householdID, ptype, status string
		pickupDate                     *time.Time
		safetyCheck                    *bool
		createdAt                      time.Time
	}

	pickups := []pickup{
		// Budi: one completed organic, one pending plastic
		{uuid.NewString(), households[0].id, "organic", "completed", &pickupDate, nil, now.Add(-5 * 24 * time.Hour)},
		{uuid.NewString(), households[0].id, "plastic", "pending", nil, nil, now.Add(-1 * 24 * time.Hour)},

		// Siti: scheduled paper
		{uuid.NewString(), households[1].id, "paper", "scheduled", &pickupDate, nil, now.Add(-3 * 24 * time.Hour)},

		// Ahmad: electronic with safety check completed
		{uuid.NewString(), households[2].id, "electronic", "completed", &pickupDate, &safetyTrue, now.Add(-6 * 24 * time.Hour)},

		// Dewi: pending organic (recent, not yet auto-canceled)
		{uuid.NewString(), households[3].id, "organic", "pending", nil, nil, now.Add(-1 * 24 * time.Hour)},

		// Dewi: canceled plastic
		{uuid.NewString(), households[3].id, "plastic", "canceled", nil, nil, now.Add(-4 * 24 * time.Hour)},

		// Riko: pending electronic (safety check true, not yet scheduled)
		{uuid.NewString(), households[4].id, "electronic", "pending", nil, &safetyTrue, now.Add(-1 * 24 * time.Hour)},
	}

	for _, p := range pickups {
		_, err := db.Exec(
			`INSERT INTO waste_pickups (id, household_id, type, status, pickup_date, safety_check, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $7) ON CONFLICT DO NOTHING`,
			p.id, p.householdID, p.ptype, p.status, p.pickupDate, p.safetyCheck, p.createdAt,
		)
		if err != nil {
			return fmt.Errorf("insert pickup: %w", err)
		}
	}
	log.Printf("  inserted %d pickups", len(pickups))

	// Payments
	// Auto-generate payments for the two completed pickups
	paidDate := now.Add(-4 * 24 * time.Hour)
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioScheme := "http"
	if getEnv("MINIO_USE_SSL", "false") == "true" {
		minioScheme = "https"
	}
	bucket := getEnv("MINIO_BUCKET", "payments")
	proofURL := fmt.Sprintf("%s://%s/%s/seed-proof.jpg", minioScheme, minioEndpoint, bucket)

	type payment struct {
		id, householdID, wasteID, status string
		amount                           float64
		paymentDate                      *time.Time
		proofURL                         *string
	}

	payments := []payment{
		// Budi organic completed → paid
		{uuid.NewString(), households[0].id, pickups[0].id, "paid", 50, &paidDate, &proofURL},
		// Ahmad electronic completed → pending (not yet confirmed)
		{uuid.NewString(), households[2].id, pickups[3].id, "pending", 100, nil, nil},
	}

	for _, p := range payments {
		_, err := db.Exec(
			`INSERT INTO payments (id, household_id, waste_id, amount, status, payment_date, proof_file_url, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) ON CONFLICT DO NOTHING`,
			p.id, p.householdID, p.wasteID, p.amount, p.status, p.paymentDate, p.proofURL,
		)
		if err != nil {
			return fmt.Errorf("insert payment: %w", err)
		}
	}
	log.Printf("  inserted %d payments", len(payments))

	return nil
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
	return db
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
