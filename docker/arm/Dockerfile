FROM easypi/alpine-arm:edge
ENTRYPOINT ["/bin/matterbridge"]

COPY . /go/src/github.com/42wim/matterbridge
RUN apk update && apk add go git gcc musl-dev ca-certificates \
        && cd /go/src/github.com/42wim/matterbridge \
        && export GOPATH=/go \
        && go get \
        && go build -x -ldflags "-X main.githash=$(git log --pretty=format:'%h' -n 1)" -o /bin/matterbridge \
        && rm -rf /go \
        && apk del --purge git go gcc musl-dev
