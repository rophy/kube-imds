.PHONY: build test test-e2e image

build:
	go build -o bin/kube-imds ./cmd/server

test:
	go test ./...

test-e2e:
	bats test/e2e/

image:
	docker build -t kube-imds:latest .
