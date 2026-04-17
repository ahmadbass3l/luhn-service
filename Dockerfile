# ── Stage 1: build ──────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go test ./... && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o luhn-service .

# ── Stage 2: minimal runtime ─────────────────────────────────────────────────
FROM scratch

COPY --from=builder /app/luhn-service /luhn-service

EXPOSE 8080

ENTRYPOINT ["/luhn-service"]
