#!/bin/bash
# The script does automatic checking on a Go package and its sub-packages, including:
# 1. gofmt         (http://golang.org/cmd/gofmt/)
# 2. goimports     (https://github.com/bradfitz/goimports)
# 3. golint        (https://github.com/golang/lint)
# 4. go vet        (http://golang.org/cmd/vet)
# 5. gosimple      (https://github.com/dominikh/go-simple)
# 6. unconvert     (https://github.com/mdempsky/unconvert)
# 7. race detector (http://blog.golang.org/race-detector)
# 8. test coverage (http://blog.golang.org/cover)
#

set -ex

# Automatic checks
for i in $(find . -name go.mod -type f -print); do
  module=$(dirname ${i})
  echo "==> ${module}"

  MODNAME=$(echo $module | sed -E -e "s/^$ROOTPATHPATTERN//" \
    -e 's,^/,,' -e 's,/v[0-9]+$,,')
  if [ -z "$MODNAME" ]; then
    MODNAME=.
  fi

  # run tests
  (cd $MODNAME &&
    echo "mode: atomic" > profile.cov && \
    env GORACE=halt_on_error=1 go test -race -covermode=atomic -coverprofile=profile.tmp ./... && \
    cat profile.tmp | tail -n +2 >> profile.cov && \
    rm profile.tmp && \
    go tool cover -func profile.cov
  )

  # check linters
  (cd $MODNAME && \
    go mod download && \
    golangci-lint run --deadline=10m --disable-all \
      --enable=gofmt \
      --enable=goimports \
      --enable=golint \
      --enable=govet \
      --enable=gosimple \
      --enable=unconvert
  )
done
