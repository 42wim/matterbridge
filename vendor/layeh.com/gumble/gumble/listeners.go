package gumble

type eventItem struct {
	parent     *Listeners
	prev, next *eventItem
	listener   EventListener
}

func (e *eventItem) Detach() {
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

// Listeners is a list of event listeners. Each attached listener is called in
// sequence when a Client event is triggered.
type Listeners struct {
	head, tail *eventItem
}

// Attach adds a new event listener to the end of the current list of listeners.
func (e *Listeners) Attach(listener EventListener) Detacher {
	item := &eventItem{
		parent:   e,
		prev:     e.tail,
		listener: listener,
	}
	if e.head == nil {
		e.head = item
	}
	if e.tail != nil {
		e.tail.next = item
	}
	e.tail = item
	return item
}

func (e *Listeners) onConnect(event *ConnectEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnConnect(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onDisconnect(event *DisconnectEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnDisconnect(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onTextMessage(event *TextMessageEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnTextMessage(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onUserChange(event *UserChangeEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnUserChange(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onChannelChange(event *ChannelChangeEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnChannelChange(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onPermissionDenied(event *PermissionDeniedEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnPermissionDenied(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onUserList(event *UserListEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnUserList(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onACL(event *ACLEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnACL(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onBanList(event *BanListEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnBanList(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onContextActionChange(event *ContextActionChangeEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnContextActionChange(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}

func (e *Listeners) onServerConfig(event *ServerConfigEvent) {
	event.Client.volatile.Lock()
	for item := e.head; item != nil; item = item.next {
		event.Client.volatile.Unlock()
		item.listener.OnServerConfig(event)
		event.Client.volatile.Lock()
	}
	event.Client.volatile.Unlock()
}
