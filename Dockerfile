# ---- build stage ----------------------------------------------------------
FROM golang:1.27 AS build

WORKDIR /src

# Download dependencies first so they are cached when only source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static, stripped binary. Migrations are embedded via //go:embed, so the final
# image only needs this one file.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/todo-api ./cmd

# ---- runtime stage -------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/todo-api /app/todo-api

# Non-root user provided by the distroless image.
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/app/todo-api"]
