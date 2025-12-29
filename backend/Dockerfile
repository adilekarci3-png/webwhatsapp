# -----------------------------
# build stage
# -----------------------------
FROM golang:1.22-alpine AS build
WORKDIR /app

RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

# tüm kaynak kodu (migrations dahil)
COPY . .

# binary build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/api

# -----------------------------
# runtime stage
# -----------------------------
FROM alpine:3.20
WORKDIR /app

# ⛳️ postgres client sadece pg_isready için
RUN apk add --no-cache ca-certificates curl postgresql-client

# server binary
COPY --from=build /app/server /app/server

# migrations klasörü
COPY --from=build /app/internal/infrastructure/persistence/postgres/migrations \
  /app/internal/infrastructure/persistence/postgres/migrations

EXPOSE 8080

# -----------------------------
# ⛳️ DB hazır olana kadar bekle
# -----------------------------
CMD sh -lc '\
  echo "Waiting for postgres at ${POSTGRES_HOST}:${POSTGRES_PORT}..."; \
  until pg_isready -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}"; do \
    sleep 2; \
  done; \
  echo "Postgres is ready. Starting backend."; \
  exec /app/server \
'
