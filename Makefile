NAME := go-modaemon
VERSION := $(shell git describe --tags --abbrev=0)
MINOR_VERSION := $(shell date +%s)
CIRCLE_SHA1 ?= $(shell git rev-parse HEAD)
CIRCLE_BRANCH ?= develop
LDFLAGS := -X 'versions.Version=$(VERSION).$(MINOR_VERSION)'
LDFLAGS += -X 'versions.Revision=$(CIRCLE_SHA1)'
LDFLAGS += -X 'versions.Branch=$(CIRCLE_BRANCH)'
PACKAGES_ALL = $(shell go list ./... | grep -v '/vendor/')
PACKAGES_MAIN = $(shell go list ./... | grep -v '/vendor/' | grep -v '/addons/')

setup:
	go get -u github.com/golang/dep/cmd/dep
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports

deps: setup
	dep ensure -v

test: deps
	go test -v ${PACKAGES_ALL}

race: deps
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

list:
	@ls -1 bin
