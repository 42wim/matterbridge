package gumble

// Channels is a map of server channels.
type Channels map[uint32]*Channel

// create adds a new channel with the given id to the collection. If a channel
// with the given id already exists, it is overwritten.
func (c Channels) create(id uint32) *Channel {
	channel := &Channel{
		ID:       id,
		Links:    Channels{},
		Children: Channels{},
		Users:    Users{},
	}
	c[id] = channel
	return channel
}

// Find returns a channel whose path (by channel name) from the server root
// channel is equal to the arguments passed. nil is returned if c does not
// containt the root channel.
func (c Channels) Find(names ...string) *Channel {
	root := c[0]
	if names == nil || root == nil {
		return root
	}
	return root.Find(names...)
}
