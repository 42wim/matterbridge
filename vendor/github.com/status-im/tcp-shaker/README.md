# TCP Checker :heartbeat:

[![Go Report Card](https://goreportcard.com/badge/github.com/tevino/tcp-shaker)](https://goreportcard.com/report/github.com/tevino/tcp-shaker)
[![GoDoc](https://godoc.org/github.com/tevino/tcp-shaker?status.svg)](https://godoc.org/github.com/tevino/tcp-shaker)
[![Build Status](https://travis-ci.org/tevino/tcp-shaker.svg?branch=master)](https://travis-ci.org/tevino/tcp-shaker)

This package is used to perform TCP handshake without ACK, which useful for TCP health checking.

HAProxy does this exactly the same, which is:

1. SYN
2. SYN-ACK
3. RST

This implementation has been running on tens of thousands of production servers for years.

## Why do I have to do this

In most cases when you establish a TCP connection(e.g. via `net.Dial`), these are the first three packets between the client and server([TCP three-way handshake][tcp-handshake]):

1. Client -> Server: SYN
2. Server -> Client: SYN-ACK
3. Client -> Server: ACK

**This package tries to avoid the last ACK when doing handshakes.**

By sending the last ACK, the connection is considered established.

However, as for TCP health checking the server could be considered alive right after it sends back SYN-ACK,

that renders the last ACK unnecessary or even harmful in some cases.

### Benefits

By avoiding the last ACK

1. Less packets better efficiency
2. The health checking is less obvious

The second one is essential because it bothers the server less.

This means the application level server will not notice the health checking traffic at all, **thus the act of health checking will not be
considered as some misbehavior of client.**

## Requirements

- Linux 2.4 or newer

There is a **fake implementation** for **non-Linux** platform which is equivalent to:

```go
conn, err := net.DialTimeout("tcp", addr, timeout)
conn.Close()
```

## Usage

```go
import "github.com/tevino/tcp-shaker"

c := NewChecker()

ctx, stopChecker := context.WithCancel(context.Background())
defer stopChecker()
go func() {
	if err := c.CheckingLoop(ctx); err != nil {
		fmt.Println("checking loop stopped due to fatal error: ", err)
	}
}()

<-c.WaitReady()

timeout := time.Second * 1
err := c.CheckAddr("google.com:80", timeout)
switch err {
case ErrTimeout:
	fmt.Println("Connect to Google timed out")
case nil:
	fmt.Println("Connect to Google succeeded")
default:
	fmt.Println("Error occurred while connecting: ", err)
}
```

## TODO

- [ ] IPv6 support (Test environment needed, PRs are welcome)

## Special thanks to contributors

- @lujjjh Added zero linger support for non-Linux platform
- @jakubgs Fixed compatibility on Android

[tcp-handshake]: https://en.wikipedia.org/wiki/Handshaking#TCP_three-way_handshake
