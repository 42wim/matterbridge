package ddp

// ----------------------------------------------------------------------
// Collection
// ----------------------------------------------------------------------

type Update map[string]interface{}
type UpdateListener interface {
	CollectionUpdate(collection, operation, id string, doc Update)
}

// Collection managed cached collection data sent from the server in a
// livedata subscription.
//
// It would be great to build an entire mongo compatible local store (minimongo)
type Collection interface {

	// FindOne queries objects and returns the first match.
	FindOne(id string) Update
	// FindAll returns a map of all items in the cache - this is a hack
	// until we have time to build out a real minimongo interface.
	FindAll() map[string]Update
	// AddUpdateListener adds a channel that receives update messages.
	AddUpdateListener(listener UpdateListener)

	// livedata updates
	added(msg Update)
	changed(msg Update)
	removed(msg Update)
	addedBefore(msg Update)
	movedBefore(msg Update)
	init()  // init informs the collection that the connection to the server has begun/resumed
	reset() // reset informs the collection that the connection to the server has been lost
}

// NewMockCollection creates an empty collection that does nothing.
func NewMockCollection() Collection {
	return &MockCache{}
}

// NewCollection creates a new collection - always KeyCache.
func NewCollection(name string) Collection {
	return &KeyCache{name, make(map[string]Update), nil}
}

// KeyCache caches items keyed on unique ID.
type KeyCache struct {
	// The name of the collection
	Name string
	// items contains collection items by ID
	items map[string]Update
	// listeners contains all the listeners that should be notified of collection updates.
	listeners []UpdateListener
	// TODO(badslug): do we need to protect from multiple threads
}

func (c *KeyCache) added(msg Update) {
	id, fields := parseUpdate(msg)
	if fields != nil {
		c.items[id] = fields
		c.notify("create", id, fields)
	}
}

func (c *KeyCache) changed(msg Update) {
	id, fields := parseUpdate(msg)
	if fields != nil {
		item, ok := c.items[id]
		if ok {
			for key, value := range fields {
				item[key] = value
			}
			c.items[id] = item
			c.notify("update", id, item)
		}
	}
}

func (c *KeyCache) removed(msg Update) {
	id, _ := parseUpdate(msg)
	if len(id) > 0 {
		delete(c.items, id)
		c.notify("remove", id, nil)
	}
}

func (c *KeyCache) addedBefore(msg Update) {
	// for keyed cache, ordered commands are a noop
}

func (c *KeyCache) movedBefore(msg Update) {
	// for keyed cache, ordered commands are a noop
}

// init prepares the collection for data updates (called when a new connection is
// made or a connection/session is resumed).
func (c *KeyCache) init() {
	// TODO start to patch up the current data with fresh server state
}

func (c *KeyCache) reset() {
	// TODO we should mark the collection but maintain it's contents and then
	// patch up the current contents with the new contents when we receive them.
	//c.items = nil
	c.notify("reset", "", nil)
}

// notify sends a Update to all UpdateListener's which should never block.
func (c *KeyCache) notify(operation, id string, doc Update) {
	for _, listener := range c.listeners {
		listener.CollectionUpdate(c.Name, operation, id, doc)
	}
}

// FindOne returns the item with matching id.
func (c *KeyCache) FindOne(id string) Update {
	return c.items[id]
}

// FindAll returns a dump of all items in the collection
func (c *KeyCache) FindAll() map[string]Update {
	return c.items
}

// AddUpdateListener adds a listener for changes on a collection.
func (c *KeyCache) AddUpdateListener(listener UpdateListener) {
	c.listeners = append(c.listeners, listener)
}

// OrderedCache caches items based on list order.
// This is a placeholder, currently not implemented as the Meteor server
// does not transmit ordered collections over DDP yet.
type OrderedCache struct {
	// ranks contains ordered collection items for ordered collections
	items []interface{}
}

func (c *OrderedCache) added(msg Update) {
	// for ordered cache, key commands are a noop
}

func (c *OrderedCache) changed(msg Update) {

}

func (c *OrderedCache) removed(msg Update) {

}

func (c *OrderedCache) addedBefore(msg Update) {

}

func (c *OrderedCache) movedBefore(msg Update) {

}

func (c *OrderedCache) init() {

}

func (c *OrderedCache) reset() {

}

// FindOne returns the item with matching id.
func (c *OrderedCache) FindOne(id string) Update {
	return nil
}

// FindAll returns a dump of all items in the collection
func (c *OrderedCache) FindAll() map[string]Update {
	return map[string]Update{}
}

// AddUpdateListener does nothing.
func (c *OrderedCache) AddUpdateListener(ch UpdateListener) {
}

// MockCache implements the Collection interface but does nothing with the data.
type MockCache struct {
}

func (c *MockCache) added(msg Update) {

}

func (c *MockCache) changed(msg Update) {

}

func (c *MockCache) removed(msg Update) {

}

func (c *MockCache) addedBefore(msg Update) {

}

func (c *MockCache) movedBefore(msg Update) {

}

func (c *MockCache) init() {

}

func (c *MockCache) reset() {

}

// FindOne returns the item with matching id.
func (c *MockCache) FindOne(id string) Update {
	return nil
}

// FindAll returns a dump of all items in the collection
func (c *MockCache) FindAll() map[string]Update {
	return map[string]Update{}
}

// AddUpdateListener does nothing.
func (c *MockCache) AddUpdateListener(ch UpdateListener) {
}

// parseUpdate returns the ID and fields from a DDP Update document.
func parseUpdate(up Update) (ID string, Fields Update) {
	key, ok := up["id"]
	if ok {
		switch id := key.(type) {
		case string:
			updates, ok := up["fields"]
			if ok {
				switch fields := updates.(type) {
				case map[string]interface{}:
					return id, Update(fields)
				default:
					// Don't know what to do...
				}
			}
			return id, nil
		}
	}
	return "", nil
}
