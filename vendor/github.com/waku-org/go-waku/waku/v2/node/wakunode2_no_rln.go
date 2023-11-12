//go:build gowaku_no_rln
// +build gowaku_no_rln

package node

import "context"

// RLNRelay is used to access any operation related to Waku RLN protocol
func (w *WakuNode) RLNRelay() RLNRelay {
	return nil
}

func (w *WakuNode) setupRLNRelay() error {
	return nil
}

func (w *WakuNode) startRlnRelay(ctx context.Context) error {
	return nil
}

func (w *WakuNode) stopRlnRelay() error {
	return nil
}
