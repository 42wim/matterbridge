package common

import "context"

// ShouldCancel returns true if the context has been cancelled and task should be aborted
func ShouldCancel(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
	}
	return false
}
