FROM ghcr.io/a-h/templ:latest AS templ
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

FROM golang:1.26-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY --from=templ /app /app
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/http cmd/http/main.go

FROM debian:bookworm-slim AS run
WORKDIR /app
COPY --from=builder /app/bin/http /app/bin/http
COPY --from=builder /app/static /app/static
COPY --from=builder /app/migrations /app/migrations
RUN adduser --disabled-password --gecos "" noroot && \
    chown -R noroot:noroot /app
USER noroot:noroot
EXPOSE 8000
CMD ["/app/bin/http"]
