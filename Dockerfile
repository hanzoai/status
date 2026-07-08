# syntax=docker/dockerfile:1

# Stage 1: Build Vite frontend (static export)
FROM node:22-alpine AS frontend
# Pin pnpm to 9.x to match the repo's lockfileVersion 9.0. `pnpm@latest` (v10)
# was non-reproducible AND treats ignored build scripts (esbuild's postinstall)
# as a FATAL error (ERR_PNPM_IGNORED_BUILDS -> exit 1), breaking the build; v9
# runs the postinstall so esbuild's platform binary is installed for vite.
RUN corepack enable && corepack prepare pnpm@9.15.9 --activate
WORKDIR /app/web/app
COPY web/app/package.json web/app/pnpm-lock.yaml* ./
# Prefer the frozen lockfile; fall back to a normal install if it has drifted.
# (The prior `... || cd web/app && ...` double-cd'd on the fallback because the
# first cd had already moved CWD — it broke every time frozen-lockfile failed.)
RUN pnpm install --frozen-lockfile || pnpm install
WORKDIR /app
COPY web/app/ web/app/
COPY web/static/brands/ web/static/brands/
RUN cd web/app && sh scripts/build.sh

# Stage 2: Build Go binary (pure Go, no CGO needed — modernc.org/sqlite)
FROM golang:1.26.4-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
COPY --from=frontend /app/web/static/ ./web/static/

# Per SCALE_STANDARD.md §2 — every Go production Dockerfile that
# emits JSON to a client builds with GOEXPERIMENT=jsonv2. Verified
# -12% time / -23% allocs on the edge POST roundtrip vs encoding/json
# v1 (json_bench_test.go in hanzoai/zip).
ARG GO_EXPERIMENT=jsonv2
ENV GOEXPERIMENT=${GO_EXPERIMENT}

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
