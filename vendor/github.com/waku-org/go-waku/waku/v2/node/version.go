package node

import (
	"fmt"
	"runtime"
)

// GitCommit is a commit hash.
var GitCommit string

// Version is the version of go-waku at the time of compilation
var Version string

type VersionInfo struct {
	Version string
	Commit  string
	System  string
	Golang  string
}

func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version: Version,
		Commit:  GitCommit,
		System:  runtime.GOARCH + "/" + runtime.GOOS,
		Golang:  runtime.Version(),
	}
}

func (v VersionInfo) String() string {
	return fmt.Sprintf("%s-%s", v.Version, v.Commit)
}
