#!/bin/bash
VERSION=$(git describe --tags)
mkdir ci/binaries
go build -x -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1)" -o ci/binaries/matterbridge-$VERSION-$GOOS-$GOARCH
cd ci
cat > deploy.json <<EOF
{
    "package": {
        "name": "Matterbridge",
        "repo": "nightly",
        "subject": "42wim"
    },
    "version": {
        "name": "$VERSION-$GOOS-$GOARCH"
    },
    "files":
        [
        {"includePattern": "binaries/(.*)", "uploadPattern":"\$1"}
        ],
    "publish": true
}
EOF

