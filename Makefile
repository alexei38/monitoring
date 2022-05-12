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
	docker run --net host $(DOCKER_IMG)

version: build
	$(BIN) --version

test:
	go test -timeout 30m -count 50 -race ./...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	$(shell go env GOPATH)/bin/golangci-lint run ./...

generate:
	go generate ./...

.PHONY: build run build-img run-img version test lint
