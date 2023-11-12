/*
Package ssdplog provides log mechanism for ssdp.
*/
package ssdplog

import "log"

var LoggerProvider = func() *log.Logger { return nil }

func Printf(s string, a ...interface{}) {
	if p := LoggerProvider; p != nil {
		if l := p(); l != nil {
			l.Printf(s, a...)
		}
	}
}
