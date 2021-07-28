FROM alpine AS builder

COPY . /go/src/matterbridge
RUN apk --no-cache add go git \
        && cd /go/src/matterbridge \
        && go build -mod vendor -ldflags "-X main.githash=$(git log --pretty=format:'%h' -n 1)" -o /bin/matterbridge

FROM alpine
RUN apk --no-cache add ca-certificates mailcap python3-dev py3-pip build-base cairo-dev cairo cairo-tools jpeg-dev zlib-dev freetype-dev lcms2-dev openjpeg-dev tiff-dev tk-dev tcl-dev \
        && pip3 install wheel lottie cairosvg
COPY --from=builder /bin/matterbridge /bin/matterbridge
RUN mkdir /etc/matterbridge \
  && touch /etc/matterbridge/matterbridge.toml \
  && ln -sf /matterbridge.toml /etc/matterbridge/matterbridge.toml
ENTRYPOINT ["/bin/matterbridge", "-conf", "/etc/matterbridge/matterbridge.toml"]
