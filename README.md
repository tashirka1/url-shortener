# url_shortener

## how to run

```bash
cp env-example .env
make air   # with autoreload
```

```bash
cp env-example .env
make up    # docker
```

## docs

[tutorial](/docs/tutorial)

## testing

```bash
go test -v ./...
```

## Benchmarking

```
wrk -t10 -c100 -d5s http://localhost:8000
```
