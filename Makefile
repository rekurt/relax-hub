.PHONY: build run test lint vet migrate-up migrate-down docker-up docker-down seed-admin clean

APP_NAME := bani-server
BUILD_DIR := ./bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	$(BUILD_DIR)/$(APP_NAME) serve

test:
	go test ./... -v

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

clean:
	rm -rf $(BUILD_DIR)
