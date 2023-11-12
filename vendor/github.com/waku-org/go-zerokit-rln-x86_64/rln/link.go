package rln

import (
	_ "github.com/waku-org/go-zerokit-rln-x86_64/libs/x86_64-pc-windows-gnu"
	_ "github.com/waku-org/go-zerokit-rln-x86_64/libs/x86_64-unknown-linux-gnu"
	_ "github.com/waku-org/go-zerokit-rln-x86_64/libs/x86_64-unknown-linux-musl"
)

/*
#cgo LDFLAGS: -lrln -ldl -lm
#cgo linux,amd64,musl,!android LDFLAGS:-L${SRCDIR}/../libs/x86_64-unknown-linux-musl
#cgo linux,amd64,!musl,!android LDFLAGS:-L${SRCDIR}/../libs/x86_64-unknown-linux-gnu
#cgo windows,amd64 LDFLAGS:-L${SRCDIR}/../libs/x86_64-pc-windows-gnu -lrln -lm -lws2_32 -luserenv
*/
import "C"
