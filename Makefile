.PHONY: build run clean tidy lint proto simulator run-sim

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

# Protobuf 代码生成
proto:
	protoc --go_out=internal/pb --go_opt=paths=source_relative -I api/proto/tbox api/proto/tbox/tbox.proto

# TBOX 模拟器
simulator:
	go build -o $(BUILD_DIR)/tbox-simulator cmd/simulator/main.go

run-sim:
	go run cmd/simulator/main.go

docker-up:
	docker-compose -f deploy/docker-compose.yaml up -d

docker-down:
	docker-compose -f deploy/docker-compose.yaml down
