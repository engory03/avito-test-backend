# ─────────────────────────────────────────────
# Stage 1: Build Go binaries
# ─────────────────────────────────────────────
FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/app/main.go -o ./docs --parseInternal --parseDependency && \
    go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server ./cmd/app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/seeder ./cmd/seed

# ─────────────────────────────────────────────
# Stage 2: Final image
# ─────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/seeder .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["/app/server"]
