# Build stage
FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/bani-server ./cmd/server

# Final stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/bani-server /app/bani-server
COPY config/config.yaml /app/config/config.yaml
COPY migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/bani-server"]
CMD ["serve"]
