-include .env
export

# Database configuration (override via .env or CLI if needed)
DB_DRIVER ?= postgres
DB_DSN    ?= postgres://postgres:postgres@localhost:5432/novabank_user_db?sslmode=disable
MIGRATIONS_DIR ?= ./migrations

# Export variables so goose picks them up automatically
export GOOSE_DRIVER=$(DB_DRIVER)
export GOOSE_DBSTRING=$(DB_DSN)
export GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR)

.PHONY: help migrate-up migrate-down migrate-reset migrate-status migrate-create

## help: Show available make commands
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## migrate-up: Apply all pending database migrations
migrate-up:
	goose up

## migrate-down: Roll back the most recent migration
migrate-down:
	goose down

## migrate-reset: Roll back all migrations to version 0
migrate-reset:
	goose reset

## migrate-status: Show the current status of all migrations
migrate-status:
	goose status

## migrate-create: Create a new SQL migration (Usage: make migrate-create name=add_transactions)
migrate-create:
	@if [ -z "$(name)" ]; then echo "Error: provide a name, e.g. make migrate-create name=add_transactions"; exit 1; fi
	goose create $(name) sql
