# Community Waste Collection API

REST API for managing household waste pickup requests and payments.

---

## Tech Stack

| Concern | Choice |
|---|---|
| Language | Go 1.25 |
| Framework | Gin |
| Database | PostgreSQL 16 |
| DB Access | sqlx + raw SQL |
| Migrations | golang-migrate |
| File Storage | MinIO (S3-compatible) |
| Rate Limiting | In-memory token bucket (`golang.org/x/time/rate`) |
| Containerization | Docker + docker-compose |

---

## Architecture

Clean Architecture with four layers:

```
internal/
  domain/       # entities, enums, repository & service interfaces — zero external deps
  repository/   # sqlx implementations of domain repository interfaces
  service/      # business logic, depends only on domain interfaces
  handler/      # Gin HTTP handlers, depend only on service/repository interfaces
  middleware/   # response helpers, rate limiter
  storage/      # MinIO/S3 adapter (implements domain.StorageService)
  worker/       # background goroutine for organic pickup auto-cancel
cmd/api/
  main.go       # DI wiring, server bootstrap, graceful shutdown
```

Dependencies flow inward: `handler → service → repository → domain`.  
All interfaces are defined in `domain/`, keeping each layer independently testable.

---

## Setup & Run

### Prerequisites
- Docker & docker-compose
- Go 1.25+ (for local development only)
- [`golang-migrate` CLI](https://github.com/golang-migrate/migrate) (for running migrations locally)

### 1. Clone and configure

```bash
git clone <repo-url>
cd community-waste-collection-go
cp .env.example .env
```

### 2. Start all services (app + PostgreSQL + MinIO)

```bash
make docker-up
# or: docker-compose up --build -d
```

The app starts on port **8080**, MinIO console on **9001**.

### 3. Run migrations

**Inside Docker (recommended):**
```bash
docker-compose exec app sh -c \
  "migrate -path /app/migrations -database 'postgres://postgres:postgres@postgres:5432/waste_collection?sslmode=disable' up"
```

**Locally (requires golang-migrate CLI):**
```bash
make migrate-up
```

### 4. Seed sample data (optional)

```bash
make seed
```

Seeds 5 households, 7 pickups (all types and statuses), and 2 payments (one paid, one pending).

### 5. Verify

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `8080` | HTTP server port |
| `APP_ENV` | `development` | Set to `production` to enable Gin release mode |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5433` | PostgreSQL port (5433 to avoid conflicts with local installs) |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `waste_collection` | Database name |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO endpoint |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO secret key |
| `MINIO_BUCKET` | `payments` | Bucket for proof-of-payment files |
| `MINIO_USE_SSL` | `false` | Use HTTPS for MinIO |

---

## API Reference

All responses follow a consistent envelope:

```json
{ "success": true, "data": { ... } }
{ "success": true, "data": [...], "meta": { "page": 1, "limit": 10, "total": 42 } }
{ "success": false, "error": "message" }
```

### Households

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/households` | Create household |
| `GET` | `/api/households` | List households (paginated) |
| `GET` | `/api/households/:id` | Get household by ID |
| `DELETE` | `/api/households/:id` | Delete household |

**POST /api/households**
```json
{ "owner_name": "John Mayer Teddy", "address": "Jl. Solo Merdeka" }
```

### Waste Pickups

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/pickups` | Create pickup request *(rate limited: 5 req/s per IP)* |
| `GET` | `/api/pickups` | List pickups (filter: `status`, `household_id`) |
| `PUT` | `/api/pickups/:id/schedule` | Schedule pickup |
| `PUT` | `/api/pickups/:id/complete` | Mark as completed (auto-creates payment) |
| `PUT` | `/api/pickups/:id/cancel` | Cancel pickup |

**POST /api/pickups**
```json
{ "household_id": "<uuid>", "type": "organic", "safety_check": null }
```
Types: `organic` `plastic` `paper` `electronic`

**PUT /api/pickups/:id/schedule**
```json
{ "pickup_date": "2026-05-10T09:00:00Z" }
```

### Payments

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/payments` | Create payment manually |
| `GET` | `/api/payments` | List payments (filter: `status`, `household_id`, `date_from`, `date_to`) |
| `PUT` | `/api/payments/:id/confirm` | Confirm payment with proof file upload |

**PUT /api/payments/:id/confirm** — `multipart/form-data`  
Field: `proof` (file — image or PDF)

### Reports

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/reports/waste-summary` | Pickup counts grouped by type and status |
| `GET` | `/api/reports/payment-summary` | Payment counts, totals by status, total revenue |
| `GET` | `/api/reports/households/:id/history` | Full pickup + payment history for a household |

---

## Business Rules

1. A household cannot create a new pickup if it has any **pending** payment.
2. A pickup can only be **scheduled** from `pending` status.
3. **Electronic** pickups require `safety_check: true` before scheduling.
4. **Organic** pickups are auto-canceled if still `pending` after 3 days (background worker, polls every minute, shuts down cleanly on SIGTERM).
5. Completing a pickup auto-generates a payment: organic/plastic/paper → **50000**, electronic → **100000**.
6. Payment confirmation requires uploading a proof file to MinIO; the URL is saved to the payment record.

---

## Development

```bash
# run locally (requires .env and running postgres/minio)
make run

# build binary
make build

# add a new migration
make migrate-create   # name of the migration

# roll back one migration
make migrate-down
```

---

## Postman Collection

Import `Community_Waste_Collection.postman_collection.json` from the repo root.  
Set the `base_url` variable to `http://localhost:8080`.
