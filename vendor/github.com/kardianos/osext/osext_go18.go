//+build go1.8

package osext

import "os"

func executable() (string, error) {
	return os.Executable()
}
