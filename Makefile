.PHONY: build run clean tidy lint

APP_NAME=iot-server
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

run:
	go run cmd/server/main.go

clean:
	rm -rf $(BUILD_DIR)

tidy:
	go mod tidy

lint:
	golangci-lint run

docker-up:
	docker-compose -f deploy/docker-compose.yaml up -d

docker-down:
	docker-compose -f deploy/docker-compose.yaml down
