.PHONY: all build install release test gofmt minimal

GIT_COMMIT := $(shell git rev-parse HEAD 2> /dev/null)
GIT_TAG := $(shell git tag --points-at HEAD 2> /dev/null | head -n 1)

# disable linking against native libc / libpthread by default;
# this can be overridden by passing CGO_ENABLED=1 to make
export CGO_ENABLED ?= 0

all: build

build:
	go build -v -ldflags "-X main.commit=$(GIT_COMMIT) -X main.version=$(GIT_TAG)"

minimal:
	go build -tags minimal -v -ldflags "-X main.commit=$(GIT_COMMIT) -X main.version=$(GIT_TAG)"

install:
	go install -v -ldflags "-X main.commit=$(GIT_COMMIT) -X main.version=$(GIT_TAG)"

release:
	goreleaser --skip-publish --clean

test:
	cd lib && go test . && go vet .
	go vet ircdog.go
	./.check-gofmt.sh

gofmt:
	./.check-gofmt.sh --fix
