.PHONY: help dev build run test clean docker-build docker-up docker-down docker-logs

help:
	@echo "Available targets:"
	@echo "  dev          - Start development environment (DB only, run app locally)"
	@echo "  dev-all      - Start full development stack (DB + app)"
	@echo "  build        - Build Go binary"
	@echo "  build-web    - Build frontend"
	@echo "  run          - Run the application locally"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start all services"
	@echo "  docker-down  - Stop all services"
	@echo "  docker-logs  - View application logs"
	@echo "  clean        - Remove build artifacts"

dev:
	docker compose up -d postgres redis

dev-all:
	docker compose up --build

build-web:
	cd web && npm ci && npm run build

build: build-web
	go build -o moltpress ./cmd/server

run: dev
	air

docker-build:
	docker build -t moltpress:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

prod-up:
	docker compose -f docker-compose.prod.yml up -d

prod-down:
	docker compose -f docker-compose.prod.yml down

prod-logs:
	docker compose -f docker-compose.prod.yml logs -f app

clean:
	rm -f moltpress
	rm -rf web/build
	docker compose down -v
