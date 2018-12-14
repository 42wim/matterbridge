#!/bin/bash
go version | grep go1.11 || exit
VERSION=$(git describe --tags)
mkdir ci/binaries
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1)" -o ci/binaries/matterbridge-$VERSION-windows-amd64.exe
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1)" -o ci/binaries/matterbridge-$VERSION-linux-amd64
GOOS=linux GOARCH=arm go build -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1)" -o ci/binaries/matterbridge-$VERSION-linux-arm
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1)" -o ci/binaries/matterbridge-$VERSION-darwin-amd64
cd ci
cat > deploy.json <<EOF
{
    "package": {
        "name": "Matterbridge",
        "repo": "nightly",
        "subject": "42wim"
    },
    "version": {
        "name": "$VERSION"
    },
    "files":
        [
        {"includePattern": "ci/binaries/(.*)", "uploadPattern":"\$1"}
        ],
    "publish": true
}
EOF

