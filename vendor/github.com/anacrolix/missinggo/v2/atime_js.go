package missinggo

import (
	"os"
	"syscall"
	"time"
)

func fileInfoAccessTime(fi os.FileInfo) time.Time {
	sys := fi.Sys().(*syscall.Stat_t)
	return time.Unix(sys.Atime, sys.AtimeNsec)
}
