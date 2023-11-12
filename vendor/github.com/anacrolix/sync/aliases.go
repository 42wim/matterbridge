package sync

import "sync"

type (
	WaitGroup = sync.WaitGroup
	Cond      = sync.Cond
	Pool      = sync.Pool
	Locker    = sync.Locker
	Once      = sync.Once
	Map       = sync.Map
)
