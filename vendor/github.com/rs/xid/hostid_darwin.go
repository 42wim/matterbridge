// +build darwin

package xid

import "golang.org/x/sys/unix"

func readPlatformMachineID() (string, error) {
	return unix.Sysctl("kern.uuid")
}
