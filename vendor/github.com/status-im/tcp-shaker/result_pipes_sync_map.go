package tcp

import "sync"

type resultPipesSyncMap struct {
	sync.Map
}

func newResultPipesSyncMap() *resultPipesSyncMap {
	return &resultPipesSyncMap{}
}

func (r *resultPipesSyncMap) popResultPipe(fd int) (chan error, bool) {
	p, exist := r.Load(fd)
	if exist {
		r.Delete(fd)
	}
	if p != nil {
		return p.(chan error), exist
	}
	return nil, exist
}

func (r *resultPipesSyncMap) deregisterResultPipe(fd int) {
	r.Delete(fd)
}

func (r *resultPipesSyncMap) registerResultPipe(fd int, pipe chan error) {
	// NOTE: the pipe should have been put back if c.fdResultPipes[fd] exists.
	r.Store(fd, pipe)
}
