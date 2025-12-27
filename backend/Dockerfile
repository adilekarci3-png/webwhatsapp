FROM golang:1.22-alpine AS build
WORKDIR /app

RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

# ✅ tüm kaynak kodu al (migrations dahil)
COPY . .

# ✅ binary build
RUN CGO_ENABLED=0 go build -o server ./cmd/api

# -----------------------------
# runtime image
# -----------------------------
FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates curl

# ✅ server binary
COPY --from=build /app/server /app/server

# ✅ migrations klasörü (container içinde artık var olacak)
COPY --from=build /app/internal/infrastructure/persistence/postgres/migrations \
  /app/internal/infrastructure/persistence/postgres/migrations

EXPOSE 8080
CMD ["/app/server"]

