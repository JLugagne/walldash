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

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY domain/ domain/
COPY pkg/ pkg/
COPY frontend/ frontend/
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Compile static binary with modernc.org/sqlite (pure Go, CGO-free)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/ha-dash ./cmd/main.go

# Stage 3: Minimal runtime image with zero external asset dependencies
FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /app/data

COPY --from=backend-builder /app/ha-dash /app/ha-dash

ENV PORT=8080 \
    DB_PATH=/app/data/ha-dash.db \
    HA_URL=http://homeassistant.local:8123 \
    HA_TOKEN=""

EXPOSE 8080

VOLUME ["/app/data"]

ENTRYPOINT ["/app/ha-dash"]
