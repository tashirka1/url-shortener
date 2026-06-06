FROM golang:1.26-alpine AS builder
WORKDIR /app
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY --chown=65532:65532 . /app
RUN templ generate
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o bin/http cmd/http/main.go

FROM debian:alpine AS run
WORKDIR /app
COPY --from=builder /app/bin/http /app/bin/http
COPY --from=builder /app/static /app/static
COPY --from=builder /app/migrations /app/migrations
RUN adduser --disabled-password --gecos "" noroot && \
    chown -R noroot:noroot /app
USER noroot:noroot
EXPOSE 8000
CMD ["/app/bin/http"]
