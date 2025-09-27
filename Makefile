.PHONY: run setup build lint test docker-up docker-down docker-logs docs

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLINT=golangci-lint

# Binary name
BINARY_NAME=backend

setup:
	$(GOCMD) mod tidy
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest

run:
	$(GOCMD) run ./cmd/api/main.go

build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/api/main.go

lint:
	$(GOLINT) run ./...

test:
	$(GOTEST) -v ./...

# Docker commands
up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f app

docs:
	swag init -g cmd/api/main.go
