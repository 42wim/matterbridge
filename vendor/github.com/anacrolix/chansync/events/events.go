package events

// Here we'll strongly-type channels to assist correct usage, if possible.

type (
	Signaled <-chan struct{}
	Done     <-chan struct{}
	Active   <-chan struct{}
	Signal   chan<- struct{}
	Acquire  chan<- struct{}
	Release  <-chan struct{}
)
