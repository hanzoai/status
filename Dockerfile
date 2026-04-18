# syntax=docker/dockerfile:1

# Stage 1: Build Vite frontend (static export)
FROM node:22-alpine AS frontend
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app
COPY web/app/package.json web/app/pnpm-lock.yaml* web/app/
RUN cd web/app && pnpm install --frozen-lockfile 2>/dev/null || cd web/app && pnpm install
COPY web/app/ web/app/
COPY web/static/brands/ web/static/brands/
RUN cd web/app && sh scripts/build.sh

# Stage 2: Build Go binary (pure Go, no CGO needed — modernc.org/sqlite)
FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
COPY --from=frontend /app/web/static/ ./web/static/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /status .

# Stage 3: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=backend /status /usr/local/bin/status
VOLUME ["/config", "/data"]
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/status"]
