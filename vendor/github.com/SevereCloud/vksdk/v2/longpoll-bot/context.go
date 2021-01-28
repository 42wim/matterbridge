package longpoll // import "github.com/SevereCloud/vksdk/v2/longpoll-bot"

import (
	"context"

	"github.com/SevereCloud/vksdk/v2/internal"
)

// TsFromContext returns the ts from context.
func TsFromContext(ctx context.Context) int {
	return ctx.Value(internal.LongPollTsKey).(int)
}
