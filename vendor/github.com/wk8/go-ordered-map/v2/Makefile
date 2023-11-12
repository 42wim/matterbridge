.DEFAULT_GOAL := all

.PHONY: all
all: test_with_fuzz lint

# the TEST_FLAGS env var can be set to eg run only specific tests
TEST_COMMAND = go test -v -count=1 -race -cover $(TEST_FLAGS)

.PHONY: test
test:
	$(TEST_COMMAND)

.PHONY: bench
bench:
	go test -bench=.

FUZZ_TIME ?= 10s

.PHONY: test_with_fuzz
test_with_fuzz:
	$(TEST_COMMAND) -fuzz=. -fuzztime=$(FUZZ_TIME)

.PHONY: fuzz
fuzz: test_with_fuzz

.PHONY: lint
lint:
	golangci-lint run
