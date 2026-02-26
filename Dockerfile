# Stage 1: Build Vue frontend
FROM node:20-alpine AS frontend
WORKDIR /app
COPY web/app/package*.json web/app/
COPY web/app/ web/app/
COPY web/app/vue.config.js web/app/
RUN cd web/app && npm ci && npm run build
# Output lands in web/static/ per vue.config.js outputDir

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
