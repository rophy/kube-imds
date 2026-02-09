.PHONY: build test image

build:
	go build -o bin/kube-imds ./cmd/server

test:
	go test ./...

image:
	docker build -t kube-imds:latest .
