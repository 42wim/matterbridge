package besticon

import (
	"os"
	"strings"
)

// PopularSites we might use for examples and testing.
var PopularSites []string

func init() {
	PopularSites = strings.Split(os.Getenv("POPULAR_SITES"), ",")
}
