#!/bin/bash
go version |grep go1.8 || exit
VERSION=$(git describe --tags)
mkdir ci/binaries
GOOS=windows GOARCH=amd64 go build -x -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1)" -o ci/binaries/matterbridge-$VERSION-win64.exe
GOOS=linux GOARCH=amd64 go build -x -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1)" -o ci/binaries/matterbridge-$VERSION-linux64
GOOS=linux GOARCH=arm go build -x -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1)" -o ci/binaries/matterbridge-$VERSION-linux-arm
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
        {"includePattern": "binaries/(.*)", "uploadPattern":"\$1"}
        ],
    "publish": true
}
EOF

