# -------------------------------------------------------------
# This makefile defines the following targets
#   - tape - builds a binary program
#   - tape-docker - build the tape image
#   - unit-test - runs the go-test based unit tests
#   - integration-test - runs the integration tests

FABRIC_VERSION = latest
INTERGATION_CASE = ANDLogic

BASE_VERSION = 0.0.2 
PREV_VERSION = 0.0.1

PROJECT_NAME = tape
DOCKERIMAGE = guoger/tape
export DOCKERIMAGE
EXTRA_VERSION ?= $(shell git rev-parse --short HEAD)
BuiltTime ?= $(shell date)
PROJECT_VERSION=$(BASE_APISERVER_VERSION)-snapshot-$(EXTRA_VERSION)

PKGNAME = github.com/guoger/$(PROJECT_NAME)
CGO_FLAGS = CGO_CFLAGS=" "
ARCH=$(shell go env GOARCH)
MARCH=$(shell go env GOOS)-$(shell go env GOARCH)
IMAGEIDS= $(shell docker images --quiet --filter=reference='$(DOCKERIMAGE)')

# defined in pkg/infra/version.go
METADATA_VAR = Version=$(BASE_VERSION)
METADATA_VAR += CommitSHA=$(EXTRA_VERSION)

GO_LDFLAGS = $(patsubst %,-X $(PROJECT_NAME)/pkg/infra.%,$(METADATA_VAR))
GO_LDFLAGS += -X '$(PROJECT_NAME)/pkg/infra.BuiltTime=$(BuiltTime)'

GO_TAGS ?=

export GO_LDFLAGS GO_TAGS FABRIC_VERSION INTERGATION_CASE

.PHONY: tape
tape:
	@echo "Building tape program...."
	go build -tags "$(GO_TAGS)" -ldflags "$(GO_LDFLAGS)" ./cmd/tape

.PHONY: unit-test
unit-test:
	@echo "Run unit test...."
	go test -v ./... --bench=. -cover

.PHONY: integration-test
integration-test:
	@echo "Run integration test"
	./test/integration-test.sh $(FABRIC_VERSION) $(INTERGATION_CASE)
