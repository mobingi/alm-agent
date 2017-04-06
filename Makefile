NAME := go-modaemon
VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.revision=$(REVISION)'
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

setup:
	go get github.com/golang/dep/...
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports

deps: setup
	dep ensure -v

test: deps
	go test -v -race ${PACKAGES}

lint: setup
	go vet ${PACKAGES}
	for pkg in ${PACKAGES}; do \
		golint -set_exit_status $$pkg || exit $$?; \
	done

fmt: setup
	goimports -v -w ${PACKAGES}

build: test
	go build -ldflags "$(LDFLAGS)" -o bin/$(NAME)

clean:
	rm bin/$(NAME)
