# webtransport-go

[![PkgGoDev](https://pkg.go.dev/badge/github.com/quic-go/webtransport-go)](https://pkg.go.dev/github.com/quic-go/webtransport-go)
[![Code Coverage](https://img.shields.io/codecov/c/github/quic-go/webtransport-go/master.svg?style=flat-square)](https://codecov.io/gh/quic-go/webtransport-go/)

webtransport-go is an implementation of the WebTransport protocol, based on [quic-go](https://github.com/quic-go/quic-go). It currently implements [draft-02](https://www.ietf.org/archive/id/draft-ietf-webtrans-http3-02.html) of the specification.

## Running a Server

```go
// create a new webtransport.Server, listening on (UDP) port 443
s := webtransport.Server{
    H3: http3.Server{Addr: ":443"},
}

// Create a new HTTP endpoint /webtransport.
http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
    conn, err := s.Upgrade(w, r)
    if err != nil {
        log.Printf("upgrading failed: %s", err)
        w.WriteHeader(500)
        return
    }
    // Handle the connection. Here goes the application logic. 
})

s.ListenAndServeTLS(certFile, keyFile)
```

Now that the server is running, Chrome can be used to establish a new WebTransport session as described in [this tutorial](https://web.dev/webtransport/).

## Running a Client

```go
var d webtransport.Dialer
rsp, conn, err := d.Dial(ctx, "https://example.com/webtransport", nil)
// err is only nil if rsp.StatusCode is a 2xx
// Handle the connection. Here goes the application logic.
```
