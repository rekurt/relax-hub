.PHONY: build run test test-hurl lint vet migrate-up migrate-down docker-up docker-down seed-admin clean swagger swagger-fmt

APP_NAME := bani-server
BUILD_DIR := ./bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

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

clean:
	rm -rf $(BUILD_DIR)

swagger:
	swag init -g cmd/server/docs.go -o docs --parseDependency --parseInternal

swagger-fmt:
	swag fmt -g cmd/server/docs.go
