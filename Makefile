.PHONY: help build up down logs clean test status shell-redis shell-engine shell-daemon shell-dashboard

help:
	@echo "LEO Mesh Simulator - Development Commands"
	@echo "=========================================="
	@echo "  make build              - Build Docker images"
	@echo "  make up                 - Start all services"
	@echo "  make down               - Stop all services"
	@echo "  make logs               - View service logs"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make test               - Run tests"
	@echo "  make status             - Show service status"
	@echo "  make engine/build       - Build orbital engine locally"
	@echo "  make dashboard/build    - Build dashboard locally"
	@echo "  make shell-*            - Open shell in a container"

build:
	docker-compose build

up:
	docker-compose up -d
	@echo "✓ Services started"
	@echo "  Orbital Engine: http://localhost:8080"
	@echo "  Routing Daemon: http://localhost:8000"
	@echo "  Dashboard: http://localhost:3000"

down:
	docker-compose down

logs:
	docker-compose logs -f

logs-engine:
	docker-compose logs -f orbital-engine

logs-daemon:
	docker-compose logs -f routing-daemon

logs-dashboard:
	docker-compose logs -f mesh-dashboard

clean:
	rm -rf orbital_engine/.bin orbital_engine/orbital-engine
	rm -rf routing_daemon/__pycache__ routing_daemon/.pytest_cache
	rm -rf mesh_dashboard/node_modules mesh_dashboard/build
	docker-compose down -v

test:
	@echo "Testing orbital engine..."
	cd orbital_engine && go mod download && go test -v ./... && go vet ./...
	@echo "Testing routing daemon..."
	cd routing_daemon && python -m py_compile app/main.py app/core/graph_manager.py app/pathfinding/connection_manager.py
	@echo "✓ All tests passed"

status:
	docker-compose ps

shell-redis:
	docker-compose exec redis redis-cli

shell-engine:
	docker-compose exec orbital-engine sh

shell-daemon:
	docker-compose exec routing-daemon bash

shell-dashboard:
	docker-compose exec mesh-dashboard sh

engine/build:
	cd orbital_engine && go mod download && go build -o orbital-engine ./cmd/engine

engine/run:
	cd orbital_engine && ./orbital-engine -redis-addr localhost:6379 -verbose

dashboard/install:
	cd mesh_dashboard && npm install

dashboard/build:
	cd mesh_dashboard && npm run build

dashboard/dev:
	cd mesh_dashboard && REACT_APP_API_URL=http://localhost:8000/api/traceroute npm start

daemon/install:
	cd routing_daemon && pip install -r requirements.txt

daemon/run:
	cd routing_daemon && REDIS_ADDR=localhost:6379 python -m app.main
