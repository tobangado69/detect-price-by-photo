# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25
ARG DISTROLESS_TAG=nonroot
ARG PLATFORM=linux/amd64

FROM --platform=${PLATFORM} golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
COPY apps/backend/go.mod apps/backend/go.sum ./
RUN go mod download
COPY apps/backend ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/backend ./cmd/

FROM --platform=${PLATFORM} gcr.io/distroless/base-debian12:${DISTROLESS_TAG}
WORKDIR /srv
COPY --from=builder /bin/backend ./backend
EXPOSE 8000
ENTRYPOINT ["/srv/backend"]
