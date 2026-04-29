BINARY=server
MIGRATIONS_PATH=migrations
DB_URL=postgres://postgres:postgres@localhost:5433/waste_collection?sslmode=disable

.PHONY: run build docker-up docker-down migrate-up migrate-down migrate-create tidy

run:
	go run ./cmd/api

build:
	go build -o $(BINARY) ./cmd/api

tidy:
	go mod tidy

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name
