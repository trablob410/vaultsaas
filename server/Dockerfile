# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod ./
RUN go mod download || true

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /valt-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /valt-migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /valt-seed ./cmd/seed

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' valt

COPY --from=builder /valt-server /usr/local/bin/valt-server
COPY --from=builder /valt-migrate /usr/local/bin/valt-migrate
COPY --from=builder /valt-seed /usr/local/bin/valt-seed

USER valt

EXPOSE 8080

CMD ["valt-server"]
