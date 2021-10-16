FROM alpine:edge as certs
RUN apk --update add ca-certificates
ARG VERSION=1.22.3
ADD https://github.com/42wim/matterbridge/releases/download/v${VERSION}/matterbridge-${VERSION}-linux-arm64 /bin/matterbridge
RUN chmod +x /bin/matterbridge

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=certs /bin/matterbridge /bin/matterbridge
ENTRYPOINT ["/bin/matterbridge"]
