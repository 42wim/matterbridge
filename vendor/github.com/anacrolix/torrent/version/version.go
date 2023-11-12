// Package version provides default versions, user-agents etc. for client identification.
package version

import (
	"fmt"
	"reflect"
	"runtime/debug"
)

var (
	DefaultExtendedHandshakeClientVersion string
	// This should be updated when client behaviour changes in a way that other peers could care
	// about.
	DefaultBep20Prefix   = "-GT0003-"
	DefaultHttpUserAgent string
	DefaultUpnpId        string
)

func init() {
	const (
		longNamespace   = "anacrolix"
		longPackageName = "torrent"
	)
	type Newtype struct{}
	var newtype Newtype
	thisPkg := reflect.TypeOf(newtype).PkgPath()
	var (
		mainPath       = "unknown"
		mainVersion    = "unknown"
		torrentVersion = "unknown"
	)
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		mainPath = buildInfo.Main.Path
		mainVersion = buildInfo.Main.Version
		// Note that if the main module is the same as this module, we get a version of "(devel)".
		for _, dep := range append(buildInfo.Deps, &buildInfo.Main) {
			if dep.Path == thisPkg {
				torrentVersion = dep.Version
			}
		}
	}
	DefaultExtendedHandshakeClientVersion = fmt.Sprintf(
		"%v %v (%v/%v %v)",
		mainPath,
		mainVersion,
		longNamespace,
		longPackageName,
		torrentVersion,
	)
	DefaultUpnpId = fmt.Sprintf("%v %v", mainPath, mainVersion)
	// Per https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent#library_and_net_tool_ua_strings
	DefaultHttpUserAgent = fmt.Sprintf(
		"%v-%v/%v",
		longNamespace,
		longPackageName,
		torrentVersion,
	)
}
