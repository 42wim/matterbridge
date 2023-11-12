# DNS Lookup Cache

[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/dnscache/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/rs/dnscache)](https://goreportcard.com/report/github.com/rs/dnscache)
[![Build Status](https://travis-ci.org/rs/dnscache.svg?branch=master)](https://travis-ci.org/rs/dnscache)
[![Coverage](http://gocover.io/_badge/github.com/rs/dnscache)](http://gocover.io/github.com/rs/dnscache)
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/dnscache)

The dnscache package provides a DNS cache layer to Go's `net.Resolver`.

# Install

Install using the "go get" command:

```
go get -u github.com/rs/dnscache
```

# Usage

Create a new instance and use it in place of `net.Resolver`. New names will be cached. Call the `Refresh` method at regular interval to update cached entries and cleanup unused ones.

```go
resolver := &dnscache.Resolver{}

// First call will cache the result
addrs, err := resolver.LookupHost(context.Background(), "example.com")

// Subsequent calls will use the cached result
addrs, err = resolver.LookupHost(context.Background(), "example.com")

// Call to refresh will refresh names in cache. If you pass true, it will also
// remove cached names not looked up since the last call to Refresh. It is a good idea
// to call this method on a regular interval.
go func() {
    t := time.NewTicker(5 * time.Minute)
    defer t.Stop()
    for range t.C {
        resolver.Refresh(true)
    }
}()
```

If you are using an `http.Transport`, you can use this cache by specifying a `DialContext` function:

```go
r := &dnscache.Resolver{}
t := &http.Transport{
    DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
        host, port, err := net.SplitHostPort(addr)
        if err != nil {
            return nil, err
        }
        ips, err := r.LookupHost(ctx, host)
        if err != nil {
            return nil, err
        }
        for _, ip := range ips {
            var dialer net.Dialer
            conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
            if err == nil {
                break
            }
        }
        return
    },
}
```

If addition to the `Refresh` method, you can `RefreshWithOptions`. This method adds an option to persist resource records
on failed lookups
```go
r := &Resolver{}
options := dnscache.ResolverRefreshOptions{}
options.ClearUnused = true
options.PersistOnFailure = false
resolver.RefreshWithOptions(options)
```
