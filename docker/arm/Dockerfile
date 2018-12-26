FROM alpine:edge as certs
RUN apk --update add ca-certificates

FROM scratch
ARG VERSION=1.12.3
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ADD https://github.com/42wim/matterbridge/releases/download/v${VERSION}/matterbridge-linux-arm /bin/matterbridge
RUN chmod +x /bin/matterbridge
ENTRYPOINT ["/bin/matterbridge"]
