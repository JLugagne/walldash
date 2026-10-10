# Home Assistant base image (multi-arch manifest with s6-overlay, bashio, tzdata).
# Pinned for build stability; the official builder overrides BUILD_FROM per
# architecture, the default below keeps local `docker build` working.
# https://developers.home-assistant.io/docs/apps/configuration#app-dockerfile
ARG BUILD_FROM=ghcr.io/home-assistant/base:3.24@sha256:8fde4170393bdf7a121ea4d6e34b417e1d0b90125948459f9f9f70cdf5d77c16

# Stage 1: Build the frontend assets
FROM node:26-alpine@sha256:143494b1da2945f061539253adc65e4f1569ddf07da2d384c022c791a9d90a4a AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
# Fail closed on a stale lockfile rather than silently resolving newer ranges.
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Stage 2: Build the Go backend binary with embedded frontend assets
FROM golang:1.27-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS backend-builder
WORKDIR /app

# BUILD_VERSION and BUILD_ARCH are provided automatically by the official
# Home Assistant builder actions. Defaults keep local builds working.
ARG BUILD_VERSION=dev
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY domain/ domain/
COPY pkg/ pkg/
COPY frontend/ frontend/
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Compile static binary with modernc.org/sqlite (pure Go, CGO-free)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -ldflags="-s -w -X main.Version=$BUILD_VERSION" -o /app/walldash ./cmd

# Stage 3: Home Assistant base image (s6-overlay is PID 1, no ENTRYPOINT/CMD here)
FROM ${BUILD_FROM}

# Declared for documentation: the builder injects the target add-on arch.
ARG BUILD_ARCH

COPY --from=backend-builder /app/walldash /usr/bin/walldash

# s6-overlay service definitions
COPY rootfs /
RUN chmod a+x /etc/services.d/walldash/run

# Unprivileged runtime user; the s6 service drops root to it (see the run script).
RUN addgroup -g 1000 -S walldash && adduser -u 1000 -S -G walldash -H -h /data walldash

LABEL \
    org.opencontainers.image.title="Walldash" \
    org.opencontainers.image.description="Touch-first 3D home automation dashboard for Home Assistant" \
    org.opencontainers.image.source="https://github.com/JLugagne/ha-dash" \
    org.opencontainers.image.licenses="MIT"

EXPOSE 8080
