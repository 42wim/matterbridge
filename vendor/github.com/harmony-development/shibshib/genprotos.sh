#!/usr/bin/env bash
mkdir -p "gen"

for dir in $(find "protocol" -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq); do
    echo "Generating files in ${dir}..."
    find "${dir}" -name '*.proto'

    protoc \
    --proto_path=protocol \
    --go_out=./gen \
    --hrpc_out=./gen \
    --hrpc_opt=hrpc-client-go \
    $(find "${dir}" -name '*.proto')
done

rsync -a -v gen/github.com/harmony-development/legato/gen/ ./gen
rm -rf ./gen/github.com

go fmt ./gen/./...

find "./gen" -name "*.go" -exec sed -i "s|github.com/harmony-development/legato|github.com/harmony-development/shibshib|g" {} \;
