package gumble

type audioEventItem struct {
	parent     *AudioListeners
	prev, next *audioEventItem
	listener   AudioListener
	streams    map[*User]chan *AudioPacket
}

func (e *audioEventItem) Detach() {
	if e.prev == nil {
		e.parent.head = e.next
	} else {
		e.prev.next = e.next
	}
	if e.next == nil {
		e.parent.tail = e.prev
	} else {
		e.next.prev = e.prev
	}
}

// AudioListeners is a list of audio listeners. Each attached listener is
// called in sequence when a new user audio stream begins.
type AudioListeners struct {
	head, tail *audioEventItem
}

// Attach adds a new audio listener to the end of the current list of listeners.
func (e *AudioListeners) Attach(listener AudioListener) Detacher {
	item := &audioEventItem{
		parent:   e,
		prev:     e.tail,
		listener: listener,
		streams:  make(map[*User]chan *AudioPacket),
	}
	if e.head == nil {
		e.head = item
	}
	if e.tail == nil {
		e.tail = item
	} else {
		e.tail.next = item
	}
	return item
}
