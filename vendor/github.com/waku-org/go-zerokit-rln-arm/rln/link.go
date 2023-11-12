package rln

import (
	_ "github.com/waku-org/go-zerokit-rln-arm/libs/aarch64-linux-android"
	_ "github.com/waku-org/go-zerokit-rln-arm/libs/aarch64-unknown-linux-gnu"
	_ "github.com/waku-org/go-zerokit-rln-arm/libs/arm-linux-androideabi"
	_ "github.com/waku-org/go-zerokit-rln-arm/libs/arm-unknown-linux-gnueabi"
	_ "github.com/waku-org/go-zerokit-rln-arm/libs/arm-unknown-linux-gnueabihf"
	_ "github.com/waku-org/go-zerokit-rln-arm/libs/armv7-linux-androideabi"
	_ "github.com/waku-org/go-zerokit-rln-arm/libs/armv7a-linux-androideabi"
)

/*
#cgo LDFLAGS: -lrln -ldl -lm
#cgo linux,arm LDFLAGS:-L${SRCDIR}/../libs/armv7-linux-androideabi
#cgo linux,arm64 LDFLAGS:-L${SRCDIR}/../libs/aarch64-unknown-linux-gnu

*/
import "C"
