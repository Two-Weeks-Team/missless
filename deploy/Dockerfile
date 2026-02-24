# ============================================================
# missless.co — Multi-stage Docker Build
# Stage 1: Go backend build
# Stage 2: Next.js static export
# Stage 3: Minimal runtime
# ============================================================

# Stage 1: Go Build
FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /server ./cmd/server

# Stage 2: Next.js Static Export
FROM node:20-alpine AS frontend
WORKDIR /app

COPY web/package*.json ./
RUN npm ci --production=false

COPY web/ .
RUN npm run build

# Stage 3: Minimal Runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Seoul

COPY --from=builder /server /server
COPY --from=frontend /app/out /web/out

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/server"]
