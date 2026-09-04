# syntax=docker/dockerfile:1

# Stage 1: Build Vite frontend (static export)
FROM ghcr.io/hanzoai/nodejs:v24.18.0 AS frontend
# Pin pnpm via corepack's packageManager field (web/app/package.json), NOT
# `pnpm@latest`: pnpm 10+ hard-errors on unapproved dependency build scripts
# (ERR_PNPM_IGNORED_BUILDS: esbuild) and dropped pnpm.onlyBuiltDependencies from
# package.json, which silently broke every Vite image build. The version lives in
# the packageManager field (pure JSON — immune to shell arg mangling that tripped
# `corepack prepare`/`npm i -g`); corepack activates it on first pnpm invocation.
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN corepack enable
WORKDIR /app
COPY web/app/package.json web/app/pnpm-lock.yaml* web/app/
RUN cd web/app && (pnpm install --frozen-lockfile || pnpm install)
COPY web/app/ web/app/
COPY web/static/brands/ web/static/brands/
RUN cd web/app && sh scripts/build.sh

# Stage 2: Build Go binary (pure Go, no CGO needed — modernc.org/sqlite)
FROM golang:1.26.5-alpine AS backend
ENV GOTOOLCHAIN=auto
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

# One directory in an empty image: the static binary and the files it reads;
# nothing else is present to run, so nothing else can be run.
FROM alpine:3.22 AS root
RUN apk add --no-cache ca-certificates tzdata && mkdir -p /config /data && chown -R 65532:65532 /config /data

FROM scratch
COPY --from=root /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=root /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=root --chown=65532:65532 /config /config
COPY --from=root --chown=65532:65532 /data /data
COPY --from=backend /status /usr/local/bin/status
VOLUME ["/config", "/data"]
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/status"]
