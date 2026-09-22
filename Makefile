.PHONY: dev-backend dev-frontend dev docker-up docker-down build build-backend build-frontend build-ubuntu

# Development — run locally
dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

# Docker
docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

docker-build:
	docker compose -f deploy/docker-compose.yml build

# Build only
build-backend:
	cd backend && go build -o lims-server ./cmd/server

build-frontend:
	cd frontend && npm run build

build: build-backend build-frontend

# Build Ubuntu deployment package
build-ubuntu:
	cd deploy && ./start.sh ubuntu