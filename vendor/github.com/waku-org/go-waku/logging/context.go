package logging

import (
	"context"

	"go.uber.org/zap"
)

var logKey = &struct{}{}

// From allows retrieving the Logger from a Context.
// Returns nil if Context does not have one.
func From(ctx context.Context) *zap.Logger {
	logger, _ := ctx.Value(logKey).(*zap.Logger)
	return logger
}

// With associates a Logger with a Context to allow passing
// a logger down the call chain.
func With(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, logKey, log)
}
