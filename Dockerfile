# Stage 1: Build Next.js frontend (static export)
FROM node:20-alpine AS frontend
WORKDIR /app
COPY web/next/package*.json web/next/
RUN cd web/next && npm ci
COPY web/next/ web/next/
COPY web/static/brands/ web/static/brands/
RUN cd web/next && sh scripts/build.sh

# Stage 2: Build Go binary (pure Go, no CGO needed — modernc.org/sqlite)
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/static/ ./web/static/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /status .

# Stage 3: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=backend /status /usr/local/bin/status
VOLUME ["/config", "/data"]
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/status"]
