NAME := go-modaemon
VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.revision=$(REVISION)'
PACKAGES_ALL = $(shell go list ./... | grep -v '/vendor/')
PACKAGES_MAIN = $(shell go list ./... | grep -v '/vendor/' | grep -v '/addons/')

setup:
	go get github.com/golang/dep/...
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports

deps: setup
	dep ensure -v

cideps: setup
	dep ensure -v --update

test: deps
	go test -v ${PACKAGES_ALL}

citest: cideps
	go test -v ${PACKAGES_ALL}

race: cideps
	go test -v -race ${PACKAGES_ALL}

lint: setup
	go vet ${PACKAGES_ALL}
	for pkg in ${PACKAGES_ALL}; do \
		golint -set_exit_status $$pkg || exit $$?; \
	done

fmt: setup
	goimports -v -w ${PACKAGES}

build: test
	go build -ldflags "$(LDFLAGS)" -o bin/$(NAME)

cibuild: race
	go build -ldflags "$(LDFLAGS)" -o bin/$(NAME)

clean:
	rm bin/$(NAME)

addon:
	cd addons/aws/; go build -ldflags "$(LDFLAGS)" -o ../../bin/$(NAME)-addon-aws
