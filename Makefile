BIN := "./bin/monitoring"
DOCKER_IMG="monitoring:develop"
REPO="github.com/alexei38/monitoring"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X $(REPO)/cmd.release="develop" \
			-X $(REPO)/cmd.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) \
			-X $(REPO)/cmd.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./

run-server: build
	$(BIN) server --config ./configs/config.yaml

run-client: build
	$(BIN) client

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run -p 9080:9080 $(DOCKER_IMG)

version: build
	$(BIN) --version

test:
	go test -race ./internal/... ./pkg/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint

generate:
	protoc -I ./proto \
			--go_out ./internal/grpc \
			--go-grpc_out ./internal/grpc \
			./proto/*.proto