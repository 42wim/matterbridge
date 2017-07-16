#!/bin/bash
VERSION=$(git describe --tags)
cd ci
mkdir binaries
go build -x -ldflags "-s -w -X main.githash=$(git log --pretty=format:'%h' -n 1) -o binaries/matterbridge-$VERSION-$GOOS-$GOARCH

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
        {"includePattern": "binaries/(*)", "uploadPattern":"\$1"}
        ],
    "publish": true
}
EOF

