# Todo Management API

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
- [Installation](#installation)
- [Environment setup](#environment-setup)
- [Database setup](#database-setup)
- [Run the application](#run-the-application)
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
   **search** (by title) and **sorting** (created date, due date, priority,
   title).
4. Organise todos with **categories / tags** (e.g. Work, Personal, Learning).
5. Deletes are **soft** — rows are marked `deleted_at` and excluded from all
   reads, never physically removed.

Cross-cutting concerns: consistent JSON response envelope, correct HTTP status
codes, request validation with per-field error messages, structured JSON
logging of every request plus application/database errors, and IP-based rate
limiting (stricter on the auth endpoints).

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
| Migrations | [`golang-migrate/migrate/v4`](https://github.com/golang-migrate/migrate) (used as a library — no CLI install needed) |
| Config | `github.com/joho/godotenv` + `os` |
| Logging | Standard library `log/slog` (structured JSON) |
| Rate limiting / request IDs | Fiber built-in `limiter` and `requestid` middleware |

## Project structure

```
todo-api/
├── cmd/
│   ├── main.go              # API entrypoint
│   └── migrate/main.go      # migration runner (up/down/version/force)
├── config/                  # env -> Config struct
├── database/
│   ├── database.go          # GORM connection + pool + auto-create DB
│   └── schema.sql           # consolidated schema (readable reference)
├── models/                  # User, Todo, Category, Date (date-only type)
├── dto/                     # request/response payloads + validation rules
├── repository/              # data access (GORM), all queries scoped by user
├── services/                # business logic, ownership rules, domain errors
├── controllers/             # thin Fiber handlers
├── middleware/              # JWT auth, rate limiter, request logger, error handler
├── routes/                  # dependency wiring + route table
├── pkg/logger/              # slog setup + GORM logger adapter
├── migrations/              # versioned .up.sql / .down.sql pairs
├── postman/                 # Postman collection
├── scripts/smoke.http       # end-to-end manual request file
├── .env.example
├── go.mod / go.sum
├── Makefile
└── README.md
```

## Installation

Prerequisites: **Go 1.27+** and a running **MySQL 8.0** server.

```bash
git clone <repository-url>
cd todo-api
go mod download
```

## Environment setup

Copy the example file and fill in real values:

```bash
cp .env.example .env      # Windows: copy .env.example .env
```

| Variable | Meaning | Example |
| --- | --- | --- |
| `APP_ENV` | `development` or `production` | `development` |
| `PORT` | HTTP port | `8080` |
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_USER` | MySQL user | `root` |
| `DB_PASSWORD` | MySQL password | `secret` |
| `DB_NAME` | Database name (auto-created if missing) | `todo_db` |
| `JWT_SECRET` | Signing secret (min 16 chars; use a long random string) | `openssl rand -hex 32` |
| `JWT_EXPIRY` | Token lifetime (Go duration) | `24h` |
| `RATE_LIMIT_MAX` | Requests per window, global | `100` |
| `RATE_LIMIT_WINDOW` | Global rate-limit window | `1m` |
| `AUTH_RATE_LIMIT_MAX` | Requests per window for `/api/auth/*` | `10` |
| `AUTH_RATE_LIMIT_WINDOW` | Auth rate-limit window | `1m` |
| `CORS_ORIGINS` | Comma-separated allowed browser origins (`*` = any) | `*` |
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` | `info` |

`.env` is git-ignored. **Never commit real credentials or secrets.**

## Database setup

The app and the migration runner both **create the database automatically** if
it does not exist (dev convenience). You only need MySQL running and valid
credentials in `.env`.

Apply the schema with the bundled migration runner (no external CLI required):

```bash
go run ./cmd/migrate up          # apply all migrations
go run ./cmd/migrate version     # show current version
go run ./cmd/migrate down        # roll back the last migration
go run ./cmd/migrate down-all    # roll back everything
```

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

Server starts on `http://localhost:8080`. Check it:

```bash
curl http://localhost:8080/health
```

## API documentation

Base URL: `http://localhost:8080`

### Auth

| Method | Path | Auth | Body | Success |
| --- | --- | --- | --- | --- |
| POST | `/api/auth/register` | — | `{ "name", "email", "password" }` | `201` |
| POST | `/api/auth/login` | — | `{ "email", "password" }` | `200` + `{ token, user }` |

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

Example: `GET /api/todos?status=pending&priority=high&sort=due_date&order=asc&page=1&limit=10`

### Categories (require auth)

| Method | Path | Body | Success |
| --- | --- | --- | --- |
| POST | `/api/categories` | `{ "name": "Work" }` | `201` |
| GET | `/api/categories` | — | `200` |

### Meta

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | service + database health |

### Postman collection

Import **`postman/TodoAPI.postman_collection.json`**. It defines a `base_url`
variable and auto-saves the JWT into a `token` collection variable after a
successful login, so the protected requests work immediately.

## Response format

Every response uses one envelope.

Success:

```json
{ "success": true, "message": "Todo created successfully", "data": { } }
```

Error:

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
`401` missing/invalid/expired token · `403` forbidden · `404` not found ·
`409` duplicate (email / category) · `429` rate limited · `500` server error.

## Notes on design decisions

- **Layered architecture** — `controllers → services → repository`. Controllers
  only parse/validate and format; services hold business rules and ownership
  checks; repositories are the only place that touches GORM.
- **Ownership** — every todo/category query includes `user_id = <current user>`.
  Accessing someone else's todo returns `404` (not `403`) so the API does not
  leak whether that ID exists.
- **Soft delete** — `Todo` embeds `gorm.DeletedAt`; GORM automatically filters
  soft-deleted rows from every read.
- **Migrations as source of truth** — `AutoMigrate` is intentionally not used;
  the schema is defined by versioned SQL in `migrations/`.
- **Minimal dependencies** — logging is stdlib `log/slog`; rate limiting and
  request IDs use Fiber's built-in middleware.
- **Graceful shutdown** — `SIGINT`/`SIGTERM` drain in-flight requests and close
  the DB pool.
