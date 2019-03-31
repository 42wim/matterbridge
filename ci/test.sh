#!/usr/bin/env bash
set -u -e -x -o pipefail

if [[ -n "${REPORT_COVERAGE+cover}" ]]; then
  # Retrieve and prepare CodeClimate's test coverage reporter.
  curl -L https://codeclimate.com/downloads/test-reporter/test-reporter-latest-linux-amd64 > ./cc-test-reporter
  chmod +x ./cc-test-reporter
  ./cc-test-reporter before-build
fi

# Run all the tests with the race detector and generate coverage.
go test -v -race -coverprofile c.out ./...

if [[ -n "${REPORT_COVERAGE+cover}" && "${TRAVIS_SECURE_ENV_VARS}" == "true" ]]; then
  # Upload test coverage to CodeClimate.
  ./cc-test-reporter after-build
fi
