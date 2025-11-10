# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY apps/ai-service/ .
RUN go mod tidy && go build -o ai-service ./cmd/

FROM gcr.io/distroless/base-debian12
WORKDIR /srv
COPY --from=builder /app/ai-service ./ai-service
ENTRYPOINT ["/srv/ai-service"]
