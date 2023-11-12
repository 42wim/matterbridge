package besticon

import (
	"errors"
	"strconv"
	"strings"
)

// SizeRange represents the desired icon dimensions
type SizeRange struct {
	Min     int
	Perfect int
	Max     int
}

var errBadSize = errors.New("besticon: bad size")

// ParseSizeRange parses a string like 60..100..200 into a SizeRange
func ParseSizeRange(s string) (*SizeRange, error) {
	parts := strings.SplitN(s, "..", 3)
	switch len(parts) {
	case 1:
		size, ok := parseSize(parts[0])
		if !ok {
			return nil, errBadSize
		}
		return &SizeRange{size, size, MaxIconSize}, nil
	case 3:
		n1, ok1 := parseSize(parts[0])
		n2, ok2 := parseSize(parts[1])
		n3, ok3 := parseSize(parts[2])
		if !ok1 || !ok2 || !ok3 {
			return nil, errBadSize
		}
		if !((n1 <= n2) && (n2 <= n3)) {
			return nil, errBadSize
		}
		return &SizeRange{n1, n2, n3}, nil
	}

	return nil, errBadSize
}

func parseSize(s string) (int, bool) {
	minSize, err := strconv.Atoi(s)
	if err != nil || minSize < MinIconSize || minSize > MaxIconSize {
		return -1, false
	}
	return minSize, true
}
