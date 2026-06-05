FROM golang:1.26-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY --chown=65532:65532 . /app
RUN go tool templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/http cmd/http/main.go

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
