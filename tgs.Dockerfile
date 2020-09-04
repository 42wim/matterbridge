FROM alpine AS builder

COPY . /go/src/github.com/42wim/matterbridge
RUN apk add \
    go \
    git \
    gcc \
    musl-dev \
  && cd /go/src/github.com/42wim/matterbridge \
  && export GOPATH=/go \
  && go get \
  && go build -x -ldflags "-X main.githash=$(git log --pretty=format:'%h' -n 1)" -o /bin/matterbridge

FROM alpine
RUN apk --no-cache add \
    ca-certificates \
    cairo \
    libjpeg-turbo \
    mailcap \
    py3-webencodings \
    python3 \
  && apk --no-cache add --virtual .compile \
    gcc \
    libffi-dev \
    libjpeg-turbo-dev \
    musl-dev \
    py3-pip \
    py3-wheel \
    python3-dev \
    zlib-dev \
  && pip3 install --no-cache-dir lottie[PNG] \
  && apk --no-cache del .compile

COPY --from=builder /bin/matterbridge /bin/matterbridge
RUN mkdir /etc/matterbridge \
  && touch /etc/matterbridge/matterbridge.toml \
  && ln -sf /matterbridge.toml /etc/matterbridge/matterbridge.toml
ENTRYPOINT ["/bin/matterbridge", "-conf", "/etc/matterbridge/matterbridge.toml"]
