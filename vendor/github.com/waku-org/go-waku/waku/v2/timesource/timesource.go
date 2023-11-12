package timesource

import (
	"context"
	"time"
)

type Timesource interface {
	Now() time.Time
	Start(ctx context.Context) error
	Stop()
}
