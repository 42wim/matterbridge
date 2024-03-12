FROM golang:1.20-bullseye AS builder

RUN apt update
RUN apt install -y gcc g++ libc6-dev

COPY . /go/src/github.com/42wim/matterbridge

WORKDIR /go/src/github.com/42wim/matterbridge

ENV GOPATH=/go

RUN go mod tidy
RUN go mod vendor
RUN go build -o /bin/matterbridge

FROM gcr.io/distroless/static-debian11

# Run ldd on binary in builder image to get a full list.
COPY --from=builder /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=builder /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=builder /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
COPY --from=builder /lib/x86_64-linux-gnu/libgcc_s.so.1 /lib/x86_64-linux-gnu/libgcc_s.so.1
COPY --from=builder /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
COPY --from=builder /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0
COPY --from=builder /lib/x86_64-linux-gnu/libresolv.so.2 /lib/x86_64-linux-gnu/libresolv.so.2
COPY --from=builder /usr/lib/x86_64-linux-gnu/libstdc++.so.6 /usr/lib/x86_64-linux-gnu/libstdc++.so.6

COPY --from=builder /bin/matterbridge /bin/matterbridge

ENTRYPOINT ["/bin/matterbridge"]
