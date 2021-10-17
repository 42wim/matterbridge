FROM alpine:edge as builder
RUN apk --update add ca-certificates curl jq
RUN VERSION=$(curl -s https://api.github.com/repos/42wim/matterbridge/releases/latest | jq -r .tag_name | cut -c 2-) && \
    curl -L -o /bin/matterbridge https://github.com/42wim/matterbridge/releases/download/v${VERSION}/matterbridge-${VERSION}-linux-arm64 && \
    chmod +x /bin/matterbridge

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /bin/matterbridge /matterbridge
ENTRYPOINT ["/matterbridge", "-conf", "/matterbridge.toml"]
