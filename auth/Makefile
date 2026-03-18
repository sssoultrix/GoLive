.PHONY: help up down logs ps build run migrate-up migrate-down db psql

COMPOSE := docker compose -f deployments/docker-compose.yml

help:
	@echo "Targets:"
	@echo "  up           Start infra + auth-api"
	@echo "  down         Stop and remove containers"
	@echo "  logs         Tail logs"
	@echo "  ps           Show container status"
	@echo "  build        Build auth-api image"
	@echo "  migrate-up   Run migrations up"
	@echo "  migrate-down Rollback migrations (1 step)"
	@echo "  db           Open Adminer in browser (http://localhost:8081)"
	@echo "  psql         Open psql inside postgres container"

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f --tail=200

ps:
	$(COMPOSE) ps

build:
	$(COMPOSE) build auth-api

migrate-up:
	$(COMPOSE) run --rm migrate up

migrate-down:
	$(COMPOSE) run --rm migrate down 1

db:
	@echo "Adminer: http://localhost:8081"
	@echo "System: PostgreSQL"
	@echo "Server: postgres"
	@echo "Username: $${POSTGRES_USER:-auth}"
	@echo "Password: $${POSTGRES_PASSWORD:-auth}"
	@echo "Database: $${POSTGRES_DB:-auth}"

psql:
	$(COMPOSE) exec postgres psql -U $${POSTGRES_USER:-auth} -d $${POSTGRES_DB:-auth}

