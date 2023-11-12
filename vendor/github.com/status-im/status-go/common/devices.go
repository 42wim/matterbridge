package common

import "runtime"

const (
	AndroidPlatform = "android"
	WindowsPlatform = "windows"
)

func OperatingSystemIs(targetOS string) bool {
	return runtime.GOOS == targetOS
}
