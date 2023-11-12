package timesource

import (
	"context"
	"time"
)

type WallClockTimeSource struct {
}

func NewDefaultClock() *WallClockTimeSource {
	return &WallClockTimeSource{}
}

func (t *WallClockTimeSource) Now() time.Time {
	return time.Now()
}

func (t *WallClockTimeSource) Start(ctx context.Context) error {
	// Do nothing
	return nil
}

func (t *WallClockTimeSource) Stop() {
	// Do nothing
}
