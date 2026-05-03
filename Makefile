.PHONY: build run test test-hurl lint vet migrate-up migrate-down docker-up docker-down seed-admin seed-demo clean swagger swagger-fmt frontend-dev frontend-build frontend-generate-api frontend-test

APP_NAME := bani-server
BUILD_DIR := ./bin
VERSION ?= dev
SWAG ?= go run github.com/swaggo/swag/cmd/swag

build:
	go build -ldflags "-X main.Version=$(VERSION) -X main.Commit=$$(git rev-parse --short HEAD) -X main.BuildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	$(BUILD_DIR)/$(APP_NAME) serve

test:
	go test ./... -v

test-hurl:
	bash tests/hurl/run_all_tests.sh

lint:
	golangci-lint run ./...

vet:
	go vet ./...

migrate-up:
	go run ./cmd/server migrate up

migrate-down:
	go run ./cmd/server migrate down

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build

seed-admin:
	@read -p "Email: " email; \
	read -p "Password: " password; \
	read -p "Name [Admin]: " name; \
	name=$${name:-Admin}; \
	go run ./cmd/server seed-admin --email="$$email" --password="$$password" --name="$$name"

seed-demo:
	go run ./cmd/server seed-demo

clean:
	rm -rf $(BUILD_DIR)

swagger:
	$(SWAG) init -g cmd/server/docs.go -o docs --parseDependency --parseInternal

swagger-fmt:
	$(SWAG) fmt -g cmd/server/docs.go

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-generate-api:
	cd frontend && npm run generate:api

frontend-test:
	cd frontend && npx vitest run
