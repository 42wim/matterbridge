.PHONY: help

help: ##@other Show this help
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

ifndef GOPATH
	$(error GOPATH not set. Please set GOPATH and make sure status-go is located at $$GOPATH/src/github.com/status-im/status-go. \
	For more information about the GOPATH environment variable, see https://golang.org/doc/code.html#GOPATH)
endif


EXPECTED_PATH=$(shell go env GOPATH)/src/github.com/status-im/doubleratchet
ifneq ($(CURDIR),$(EXPECTED_PATH))
define NOT_IN_GOPATH_ERROR

Current dir is $(CURDIR), which seems to be different from your GOPATH.
Please, build status-go from GOPATH for proper build.
  GOPATH       = $(shell go env GOPATH)
  Current dir  = $(CURDIR)
  Expected dir = $(EXPECTED_PATH))
See https://golang.org/doc/code.html#GOPATH for more info

endef
$(error $(NOT_IN_GOPATH_ERROR))
endif

GOBIN=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))build/bin
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# This is a code for automatic help generator.
# It supports ANSI colors and categories.
# To add new item into help output, simply add comments
# starting with '##'. To add category, use @category.
GREEN  := $(shell echo "\e[32m")
WHITE  := $(shell echo "\e[37m")
YELLOW := $(shell echo "\e[33m")
RESET  := $(shell echo "\e[0m")
HELP_FUN = \
		   %help; \
		   while(<>) { push @{$$help{$$2 // 'options'}}, [$$1, $$3] if /^([a-zA-Z0-9\-]+)\s*:.*\#\#(?:@([a-zA-Z\-]+))?\s(.*)$$/ }; \
		   print "Usage: make [target]\n\n"; \
		   for (sort keys %help) { \
			   print "${WHITE}$$_:${RESET}\n"; \
			   for (@{$$help{$$_}}) { \
				   $$sep = " " x (32 - length $$_->[0]); \
				   print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
			   }; \
			   print "\n"; \
		   }

setup: lint-install ##@other Prepare project for first build

lint-install:
	@# The following installs a specific version of golangci-lint, which is appropriate for a CI server to avoid different results from build to build
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b $(GOPATH)/bin v1.9.1

lint: ##@other Run linter
	@echo "lint"
	@golangci-lint run ./...
