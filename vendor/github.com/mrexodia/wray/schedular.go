package wray

import "time"

type Schedular interface {
	wait(time.Duration, func())
	delay() time.Duration
}

type ChannelSchedular struct {
}

func (self ChannelSchedular) wait(delay time.Duration, callback func()) {
	go func() {
		time.Sleep(delay)
		callback()
	}()
}

func (self ChannelSchedular) delay() time.Duration {
	return (1 * time.Minute)
}
