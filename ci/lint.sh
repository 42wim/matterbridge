#!/usr/bin/env bash
set -u -e -x -o pipefail

if [[ -n "${GOLANGCI_VERSION-}" ]]; then
  # Retrieve the golangci-lint linter binary.
  curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ${GOPATH}/bin ${GOLANGCI_VERSION}
fi

# Run the linter.
golangci-lint run

# if [[ "${GO111MODULE-off}" == "on" ]]; then
#   # If Go modules are active then check that dependencies are correctly maintained.
#   go mod tidy
#   go mod vendor
#   git diff --exit-code --quiet || (echo "Please run 'go mod tidy' to clean up the 'go.mod' and 'go.sum' files."; false)
# fi
