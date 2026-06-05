# url_shortener

## how to run
```
cp env-example .env
make up
```

or
```
cp env-example .env
go tool air
```

## docs
[open docs](/docs)

## use cases
register user
login user
create url
list urls
redirect url

## wrk
```
wrk -c 100 -t 10 -d 10 --latency http://localhost:8000/link/create-link
```

## escape analysis
```
go build -gcflags=-m -o bin/http cmd/http/main.go 2>&1 | grep hashtriemap.go
```

## build small binary
```
go build -trimpath -ldflags="-s -w" -o bin/http cmd/http/main.go
```

## testing
```
go test -v
go test -v -coverprofile cover.out ./...
go tool cover -html=cover.out
```
