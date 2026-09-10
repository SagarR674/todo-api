# Todo Management API - developer tasks
#
# On Windows, run these with `make <target>` if GNU Make is installed, or copy
# the command shown for each target. All targets read configuration from .env.

.PHONY: help tidy run build test migrate-up migrate-down migrate-down-all migrate-version migrate-create

help:
	@echo "targets:"
	@echo "  tidy             - go mod tidy"
	@echo "  run              - run the API server (go run ./cmd)"
	@echo "  build            - build the server binary into ./bin"
	@echo "  test             - run unit tests"
	@echo "  migrate-up       - apply all pending migrations"
	@echo "  migrate-down     - roll back the last migration"
	@echo "  migrate-down-all - roll back every migration"
	@echo "  migrate-version  - print the current schema version"
	@echo "  migrate-create name=<n> - scaffold a new migration pair"

tidy:
	go mod tidy

run:
	go run ./cmd

build:
	go build -o bin/todo-api ./cmd

test:
	go test ./...

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-down-all:
	go run ./cmd/migrate down-all

migrate-version:
	go run ./cmd/migrate version

# Example: make migrate-create name=add_todo_tags
migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=<n>" && exit 1)
	@next=$$(printf "%06d" $$(( $$(ls migrations 2>/dev/null | sed -n 's/^\([0-9]\{6\}\).*/\1/p' | sort -n | tail -1 | sed 's/^0*//' ) + 1 )) ); \
	touch migrations/$${next}_$(name).up.sql migrations/$${next}_$(name).down.sql; \
	echo "created migrations/$${next}_$(name).{up,down}.sql"
