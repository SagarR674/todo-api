# Todo Management API - developer tasks
#
# On Windows use `make <target>` with GNU Make installed, or run the command
# shown for each target directly. All targets read configuration from .env.

.PHONY: help tidy run build test test-unit test-integration cover \
        vet staticcheck lint fmt check \
        migrate-up migrate-down migrate-down-all migrate-version migrate-create \
        docker-up docker-down

help:
	@echo "targets:"
	@echo "  run              - run the API server (go run ./cmd)"
	@echo "  build            - build the server binary into ./bin"
	@echo "  test             - run every test (needs a test MySQL; see README)"
	@echo "  test-unit        - run only the fast unit tests (no database)"
	@echo "  test-integration - run only the HTTP integration tests"
	@echo "  cover            - run tests with a coverage report"
	@echo "  check            - fmt check + vet + staticcheck + tests"
	@echo "  lint             - vet + staticcheck"
	@echo "  migrate-up       - apply all pending migrations"
	@echo "  migrate-down     - roll back the last migration"
	@echo "  migrate-version  - print the current schema version"
	@echo "  migrate-create name=<n> - scaffold a new migration pair"
	@echo "  docker-up        - build and start the full stack with docker compose"

tidy:
	go mod tidy

run:
	go run ./cmd

build:
	go build -o bin/todo-api ./cmd

test:
	go test ./...

test-unit:
	go test ./config/... ./utils/... ./models/... ./dto/... ./services/...

test-integration:
	REQUIRE_DB=1 go test ./test/...

cover:
	go test -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

fmt:
	gofmt -w .

vet:
	go vet ./...

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

lint: vet staticcheck

check:
	@test -z "$$(gofmt -l .)" || (echo "run 'make fmt'" && gofmt -l . && exit 1)
	$(MAKE) vet
	$(MAKE) staticcheck
	$(MAKE) test

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

docker-up:
	docker compose up --build

docker-down:
	docker compose down
