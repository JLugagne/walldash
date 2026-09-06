# Stage 1: Build the frontend assets
FROM node:26-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci || npm install

COPY frontend/ ./
RUN npm run build

# Stage 2: Build the Go backend binary with embedded frontend assets
FROM golang:1.27-alpine AS backend-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY domain/ domain/
COPY pkg/ pkg/
COPY frontend/ frontend/
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Compile static binary with modernc.org/sqlite (pure Go, CGO-free)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/walldash ./cmd/main.go

# Stage 3: Scratch image — copy only the binary and runtime deps
FROM alpine:3.21 AS runtime-deps
RUN apk add --no-cache ca-certificates tzdata && mkdir -p /app/data

FROM scratch
WORKDIR /app

COPY --from=runtime-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=runtime-deps /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=runtime-deps /app/data /app/data

COPY --from=backend-builder /app/walldash /app/walldash

ENV PORT=8080 \
    DB_PATH=/app/data/walldash.db \
    HA_URL=http://homeassistant.local:8123 \
    HA_TOKEN=""

EXPOSE 8080

VOLUME ["/app/data"]

ENTRYPOINT ["/app/walldash"]
