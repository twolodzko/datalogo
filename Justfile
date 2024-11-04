test:
	go test -count=1 ./...

cov:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

staticcheck:
	staticcheck ./...

repl:
	@ go run .

build:
	go build -ldflags="-s -w" .

clean:
	go mod tidy
	go fmt
	rm -rf *.out *.html *.prof *.test
	go clean -testcache
