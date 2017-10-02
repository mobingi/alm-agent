NAME := alm-agent
VERSION := $(shell cat versions/version_base)
DOCKER_REVISION := $(shell cat DOCKER_REVISION)
MINOR_VERSION := $(shell date +%s)
CIRCLE_SHA1 ?= $(shell git rev-parse HEAD)
CIRCLE_BRANCH ?= develop
LDFLAGS := -X 'github.com/mobingi/alm-agent/versions.Version=$(VERSION).$(MINOR_VERSION)'
LDFLAGS += -X 'github.com/mobingi/alm-agent/versions.Revision=$(CIRCLE_SHA1)'
LDFLAGS += -X 'github.com/mobingi/alm-agent/versions.Branch=$(CIRCLE_BRANCH)'
LDFLAGS += -X 'main.RollbarToken=$(ROLLBAR_CLIENT_TOKEN)'
PACKAGES_ALL = $(shell go list ./...')
PACKAGES_MAIN = $(shell go list ./... | grep -v '/addons/')

vendor.conf: setup
	curl -LO https://raw.githubusercontent.com/moby/moby/${DOCKER_REVISION}/vendor.conf
	echo >> vendor.conf
	echo github.com/docker/docker ${DOCKER_REVISION} >> vendor.conf
	cat vendor_append.conf >> vendor.conf

setup:
	go get -u github.com/golang/dep/cmd/dep
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports
	go get -u github.com/jteeuwen/go-bindata/...
	go get github.com/BurntSushi/toml/cmd/tomlv

deps:
	dep ensure -v

.PHONY: bindata
bindata:
	tomlv _data/*.toml
	go-bindata -o ./bindata/bindata.go -pkg bindata -nometadata ./_data/

verifydata:
	tomlv _data/*.toml
	go-bindata -o ./checkbin -pkg bindata -nometadata ./_data/
	diff ./checkbin ./bindata/bindata.go > /dev/null
	rm ./checkbin

test: deps
	go test -v ${PACKAGES_ALL} -cover

race: deps
	go test -v -race ${PACKAGES_ALL} -cover

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

version:
	@echo ${VERSION}.${MINOR_VERSION}
