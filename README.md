# Todo Management API

[![CI](https://github.com/SagarR674/todo-api/actions/workflows/ci.yml/badge.svg)](https://github.com/SagarR674/todo-api/actions/workflows/ci.yml)

A backend REST API for a Todo Management System, built with **Go** and the
**Fiber** web framework, backed by **MySQL**. It provides user registration and
login with **JWT authentication**, and full CRUD for per-user todos with
filtering, search, sorting, pagination, categories/tags, soft delete, structured
logging and rate limiting.

> **Stack note:** The assignment brief's sample environment block lists Postgres
> values, but Section 2 mandates **MySQL** as the database. This project uses
> MySQL, as required by the stack.

---

## Table of contents

- [Project overview](#project-overview)
- [Technology stack](#technology-stack)
- [Project structure](#project-structure)
- [Quick start with Docker](#quick-start-with-docker)
- [Installation (local Go)](#installation-local-go)
- [Environment setup](#environment-setup)
- [Database setup](#database-setup)
- [Run the application](#run-the-application)
- [Testing](#testing)
- [API documentation](#api-documentation)
- [Response format](#response-format)
- [Notes on design decisions](#notes-on-design-decisions)

---

## Project overview

The API lets a user:

1. Register and log in. Passwords are hashed with **bcrypt**; login returns a
   signed **JWT**.
2. Manage their own todos (create, list, read, update, change status, delete).
   Every todo operation is scoped to the authenticated user — you can never see
   or touch another user's todo.
3. List todos with **pagination**, **filtering** (status, priority, category),
   **search** (by title) and **sorting** (created date, updated date, due date,
   priority, title).
4. Organise todos with **categories / tags** (e.g. Work, Personal, Learning).
5. Deletes are **soft** — rows are marked `deleted_at` and excluded from all
   reads, never physically removed.

Cross-cutting concerns: consistent JSON response envelope, correct HTTP status
codes, request validation with per-field error messages, structured JSON
logging of every request plus application/database errors, IP-based rate
limiting (stricter on the auth endpoints), CORS, request IDs and graceful
shutdown.

## Technology stack

| Area | Choice |
| --- | --- |
| Language | Go 1.27 |
| Web framework | [Fiber v2](https://github.com/gofiber/fiber) |
| Database | MySQL 8.0 |
| ORM | [GORM](https://gorm.io) + `gorm.io/driver/mysql` |
| Auth | JWT — [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt) (HS256) |
| Password hashing | `golang.org/x/crypto/bcrypt` |
| Validation | [`go-playground/validator/v10`](https://github.com/go-playground/validator) |
| Migrations | [`golang-migrate/migrate/v4`](https://github.com/golang-migrate/migrate), SQL **embedded** in the binary |
| Config | `github.com/joho/godotenv` + `os` |
| Logging | Standard library `log/slog` (structured JSON) |
| Rate limiting / CORS / request IDs | Fiber built-in middleware |
| Tests | stdlib `testing` + `net/http/httptest`; CI on GitHub Actions with a MySQL service |

## Project structure

```
todo-api/
├── cmd/
│   ├── main.go              # API entrypoint
│   └── migrate/main.go      # migration CLI (up/down/version/force)
├── config/                  # env -> Config struct (+ unit tests)
├── database/
│   ├── database.go          # GORM connection + pool + auto-create DB
│   ├── migrate.go           # run embedded migrations (used by app, CLI and tests)
│   └── schema.sql           # consolidated schema (readable reference)
├── models/                  # User, Todo, Category, Date (date-only type) (+ tests)
├── dto/                     # request/response payloads + validation rules (+ tests)
├── repository/              # data access (GORM), all queries scoped by user
├── services/                # business logic, ownership rules, domain errors (+ tests)
├── controllers/             # thin Fiber handlers
├── middleware/              # JWT auth, rate limiter, CORS, request logger, error handler
├── routes/                  # route table
├── server/                  # assembles the Fiber app (shared by main + tests)
├── pkg/logger/              # slog setup + GORM logger adapter
├── migrations/              # versioned .up.sql / .down.sql pairs (embedded via //go:embed)
├── test/                    # black-box HTTP integration tests
├── postman/                 # Postman collection
├── scripts/smoke.http       # end-to-end manual request file
├── .github/workflows/ci.yml # build + vet + staticcheck + race tests
├── Dockerfile               # multi-stage, distroless, non-root
├── docker-compose.yml       # api + mysql, one command
├── .env.example
├── Makefile
└── README.md
```

## Quick start with Docker

The fastest way to run everything (no Go or MySQL install needed):

```bash
docker compose up --build
```

This starts MySQL, waits for it to be healthy, applies migrations automatically
(`AUTO_MIGRATE=true`), and serves the API on <http://localhost:8080>.

```bash
curl http://localhost:8080/health
```

Override secrets/ports with a `.env` file or shell environment (see
`docker-compose.yml` for the variables it reads).

## Installation (local Go)

Prerequisites: **Go 1.27+** and a running **MySQL 8.0** server.

```bash
git clone https://github.com/SagarR674/todo-api.git
cd todo-api
go mod download
```

## Environment setup

Copy the example file and fill in real values:

```bash
cp .env.example .env      # Windows: copy .env.example .env
```

| Variable | Meaning | Default |
| --- | --- | --- |
| `APP_ENV` | `development` or `production` | `development` |
| `PORT` | HTTP port | `8080` |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` | MySQL connection | `localhost` / `3306` / `root` / — |
| `DB_NAME` | Database name (auto-created if missing) | `todo_db` |
| `AUTO_MIGRATE` | Apply pending migrations on startup | `false` |
| `JWT_SECRET` | Signing secret, min 16 chars (use `openssl rand -hex 32`) | — (required) |
| `JWT_EXPIRY` | Token lifetime (Go duration) | `24h` |
| `RATE_LIMIT_MAX` / `RATE_LIMIT_WINDOW` | Global rate limit | `100` / `1m` |
| `AUTH_RATE_LIMIT_MAX` / `AUTH_RATE_LIMIT_WINDOW` | Stricter limit for `/api/auth/*` | `10` / `1m` |
| `CORS_ORIGINS` | Comma-separated allowed browser origins (`*` = any) | `*` |
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` | `info` |

`.env` is git-ignored. **Never commit real credentials or secrets.**

## Database setup

The app and the migration CLI both **create the database automatically** if it
does not exist. Migrations are embedded in the binary, so no external CLI or
files are required:

```bash
go run ./cmd/migrate up          # apply all migrations
go run ./cmd/migrate version     # show current schema version
go run ./cmd/migrate down        # roll back the last migration
go run ./cmd/migrate down-all    # roll back everything
```

Or let the server run them on boot: set `AUTO_MIGRATE=true`.

Prefer raw SQL? `database/schema.sql` is a consolidated, readable copy:

```bash
mysql -u root -p < database/schema.sql
```

## Run the application

```bash
go run ./cmd
# or build a binary:
go build -o bin/todo-api ./cmd && ./bin/todo-api
```

Server starts on <http://localhost:8080>. Check it:

```bash
curl http://localhost:8080/health
```

## Testing

The suite has two layers:

| Layer | Location | Needs a database? |
| --- | --- | --- |
| Unit tests | `config/`, `utils/`, `models/`, `dto/`, `services/` | No |
| HTTP integration tests | `test/` | Yes — its own MySQL schema |

```bash
make test-unit          # fast, no database
make test-integration   # full HTTP flow against MySQL
make test               # everything
make cover              # everything + coverage summary
```

The integration suite uses a **separate schema** (`TEST_DB_NAME`, default
`<DB_NAME>_test`) so it can create and truncate tables freely. It reads
`TEST_DB_*` environment variables and falls back to the `DB_*` values. If the
database is unreachable the suite **skips** itself; set `REQUIRE_DB=1` (as CI
does) to make an unreachable database a failure instead.

```bash
# example: run integration tests against a local MySQL
TEST_DB_PASSWORD=secret REQUIRE_DB=1 go test ./test/...
```

CI (`.github/workflows/ci.yml`) runs `gofmt` check, `go vet`, `staticcheck`, and
`go test -race` against a MySQL 8.0 service container on every push and PR.

## API documentation

Base URL: `http://localhost:8080`

### Auth

| Method | Path | Auth | Body | Success |
| --- | --- | --- | --- | --- |
| POST | `/api/auth/register` | — | `{ "name", "email", "password" }` | `201` |
| POST | `/api/auth/login` | — | `{ "email", "password" }` | `200` + `{ token, user }` |
| GET | `/api/auth/me` | Bearer | — | `200` + current user |

`password` must be at least 8 characters. `email` must be unique and valid.

### Todos (require `Authorization: Bearer <token>`)

| Method | Path | Body | Notes |
| --- | --- | --- | --- |
| POST | `/api/todos` | create payload | `201` |
| GET | `/api/todos` | — | list; supports query params below |
| GET | `/api/todos/:id` | — | `404` if not yours |
| PUT | `/api/todos/:id` | partial update payload | any subset of fields |
| PATCH | `/api/todos/:id/status` | `{ "status" }` | |
| DELETE | `/api/todos/:id` | — | soft delete |

**Create / update payload**

```json
{
  "title": "Complete backend assignment",
  "description": "Develop Todo APIs using Golang and Fiber",
  "status": "pending",
  "priority": "high",
  "due_date": "2026-09-15",
  "category_ids": [1, 2]
}
```

- `status` ∈ `pending | in_progress | completed` (default `pending`)
- `priority` ∈ `low | medium | high` (default `medium`)
- `due_date` format `YYYY-MM-DD`
- `category_ids` must reference categories you own

**List query parameters** — `GET /api/todos`

| Param | Example | Description |
| --- | --- | --- |
| `page` | `1` | page number (default 1) |
| `limit` | `10` | page size (default 10, max 100) |
| `status` | `pending` | filter by status |
| `priority` | `high` | filter by priority |
| `search` | `backend` | case-insensitive match on title |
| `category` | `Work` | filter by category name |
| `sort` | `due_date` | one of `created_at`, `updated_at`, `due_date`, `priority`, `title` |
| `order` | `desc` | `asc` or `desc` (default `desc`) |

Response `data`:

```json
{
  "items": [ /* todos */ ],
  "pagination": { "page": 1, "limit": 10, "total": 42, "total_pages": 5, "has_next": true, "has_prev": false }
}
```

Example: `GET /api/todos?status=pending&priority=high&sort=due_date&order=asc&page=1&limit=10`

### Categories (require auth)

| Method | Path | Body | Success |
| --- | --- | --- | --- |
| POST | `/api/categories` | `{ "name": "Work" }` | `201` (`409` on duplicate) |
| GET | `/api/categories` | — | `200` |

### Meta

| Method | Path | Description |
| --- | --- | --- |
| GET | `/` | API name, version and endpoint map |
| GET | `/health` | service + database health |

### Postman collection

Import **`postman/TodoAPI.postman_collection.json`**. It defines a `base_url`
variable and auto-saves the JWT into a `token` collection variable after a
successful login, so the protected requests work immediately.

## Response format

Every response uses one envelope.

```json
{ "success": true, "message": "Todo created successfully", "data": { } }
```

```json
{ "success": false, "message": "Todo not found" }
```

Validation error (`400`):

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": { "title": "title is required", "priority": "must be one of: low, medium, high" }
}
```

HTTP status codes: `200` OK · `201` Created · `400` validation/bad input ·
`401` missing/invalid/expired token · `404` not found · `409` duplicate
(email / category) · `429` rate limited · `500` server error.

## Notes on design decisions

- **Layered architecture** — `controllers → services → repository`. Controllers
  only parse/validate and format; services hold business rules and ownership
  checks; repositories are the only place that touches GORM. Services depend on
  small interfaces (`services.UserRepo`, `services.TodoRepo`, …) so business
  logic is unit-tested with in-memory fakes.
- **Ownership** — every todo/category query includes `user_id = <current user>`.
  Accessing another user's todo returns **`404`, not `403`**, so the API does not
  leak whether that ID exists (a common security practice). The brief lists
  `403` as an example status code, not a requirement.
- **Soft delete** — `Todo` embeds `gorm.DeletedAt`; GORM automatically filters
  soft-deleted rows from every read. The row and its `deleted_at` timestamp
  remain in the table.
- **Migrations as source of truth** — GORM `AutoMigrate` is intentionally not
  used; the schema is versioned SQL in `migrations/`, embedded into the binary
  via `//go:embed` and run through `golang-migrate`.
- **Minimal dependencies** — logging is stdlib `log/slog`; rate limiting, CORS
  and request IDs use Fiber's built-in middleware.
- **Same wiring everywhere** — `server.New()` builds the app for both the main
  binary and the integration tests, so tests exercise the real middleware stack.
- **Graceful shutdown** — `SIGINT`/`SIGTERM` drain in-flight requests and close
  the DB pool.
