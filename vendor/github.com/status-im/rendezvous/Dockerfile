FROM golang:1.17.0-alpine as builder

RUN apk add --no-cache gcc musl-dev linux-headers

RUN mkdir -p /go/src/github.com/status-im/rendezvous
ADD . /go/src/github.com/status-im/rendezvous
RUN cd /go/src/github.com/status-im/rendezvous && go build -o rendezvous ./cmd/server/

FROM alpine:latest

RUN apk add --no-cache ca-certificates bash

COPY --from=builder /go/src/github.com/status-im/rendezvous/rendezvous /usr/local/bin/rendezvous
