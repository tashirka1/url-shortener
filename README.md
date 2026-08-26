# url-shortener

## how to run

Development
```bash
cp env-example .env
make dev   # with autoreload
```

Production Docker
```bash
cp env-example .env
make up    # docker
```

Production Binary
```bash
sudo apt-get update && sudo apt-get install -y --no-install-recommends build-essential libsqlite3-dev
cp env-example .env
make build-bin    # binary
```

## docs

[tutorial](/docs/tutorial)

## testing

```bash
go test -v ./...
```

## Benchmarking (RPS)

Pre-populate DB for SELECT/UPDATE benchmarks:
```bash
for i in $(seq 1 1000); do curl -s "http://localhost:8000/rps/templ-page-insert?payload=prefill" > /dev/null; done
```

Run benchmarks:
```bash
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-text
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-json
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-templ-page
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-insert?payload=bench'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-select-join?limit=15'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-update'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-select-simple?limit=15's
```
