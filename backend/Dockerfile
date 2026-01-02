# --- build stage ---
FROM golang:1.24-alpine AS build
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api

# --- runtime stage ---
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates curl

COPY --from=build /app/server /app/server

# ✅ migrations'ı da kopyala (kritik)
COPY --from=build /app/internal/infrastructure/persistence/postgres/migrations \
  /app/internal/infrastructure/persistence/postgres/migrations

RUN chmod +x /app/server
CMD ["/app/server"]
