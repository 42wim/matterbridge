package dht

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anacrolix/dht/v2/krpc"
)

var TransactionTimeout = errors.New("transaction timed out")

// Transaction keeps track of a message exchange between nodes, such as a
// query message and a response message.
type Transaction struct {
	onResponse func(krpc.Msg)
}

func (t *Transaction) handleResponse(m krpc.Msg) {
	t.onResponse(m)
}

const defaultMaxQuerySends = 1

func transactionSender(
	ctx context.Context,
	send func() error,
	resendDelay func() time.Duration,
	maxSends int,
) error {
	var delay time.Duration
	sends := 0
	for sends < maxSends {
		select {
		case <-time.After(delay):
			err := send()
			sends++
			if err != nil {
				return fmt.Errorf("send %d: %w", sends, err)
			}
			delay = resendDelay()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
