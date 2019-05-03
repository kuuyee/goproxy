.PHONY: build generate image clean test

export GO111MODULE=on
export CGO_ENABLED=0

all: build

build: generate
	@go build -o bin/goproxy -ldflags "-s -w" .

generate:
	@go generate
	@go mod tidy

image:
	@docker build -t goproxy/goproxy .

test: generate
	@go test -v `(go list ./... | grep "pkg/proxy")`

clean:
	@git clean -f -d -X
