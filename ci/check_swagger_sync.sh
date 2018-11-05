#!/bin/bash

$GOPATH/bin/swag init --generalInfo bridge/api/api.go --dir . 2> /dev/null

# When regenerated, only a single date line changes.
# A minimal one-line diff shows as two operations (addition + subtraction),
# with 7 lines in the diff output
if [ "$(git diff --unified=0 | wc -l)" -gt 7 ]; then
  exit 1
fi
