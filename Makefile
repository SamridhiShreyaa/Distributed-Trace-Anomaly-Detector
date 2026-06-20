.PHONY: test build demo-up demo-down inject-regression clean

test:
	go test ./... -v -race -coverprofile=coverage.out

build:
	go build -o bin/detector ./cmd/detector

demo-up:
	docker compose up -d

demo-down:
	docker compose down

inject-regression:
	go run demo/load/inject_regression.go --service=cart --latency=300ms --duration=2m

clean:
	rm -rf bin/
	rm -f coverage.out
	go clean -testcache

.DEFAULT_GOAL := build
