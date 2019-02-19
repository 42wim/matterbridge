package ddp

// ------------------------------------------------------------
// DDP Messages
//
// Go structs representing DDP raw messages ready for JSON
// encoding.
// ------------------------------------------------------------

// Message contains the common fields that all DDP messages use.
type Message struct {
	Type string `json:"msg"`
	ID   string `json:"id,omitempty"`
}

// Connect represents a DDP connect message.
type Connect struct {
	Message
	Version string   `json:"version"`
	Support []string `json:"support"`
	Session string   `json:"session,omitempty"`
}

// NewConnect creates a new connect message
func NewConnect() *Connect {
	return &Connect{Message: Message{Type: "connect"}, Version: "1", Support: []string{"1"}}
}

// NewReconnect creates a new connect message with a session ID to resume.
func NewReconnect(session string) *Connect {
	c := NewConnect()
	c.Session = session
	return c
}

// Ping represents a DDP ping message.
type Ping Message

// NewPing creates a new ping message with optional ID.
func NewPing(id string) *Ping {
	return &Ping{Type: "ping", ID: id}
}

// Pong represents a DDP pong message.
type Pong Message

// NewPong creates a new pong message with optional ID.
func NewPong(id string) *Pong {
	return &Pong{Type: "pong", ID: id}
}

// Method is used to send a remote procedure call to the server.
type Method struct {
	Message
	ServiceMethod string        `json:"method"`
	Args          []interface{} `json:"params"`
}

// NewMethod creates a new method invocation object.
func NewMethod(id, serviceMethod string, args []interface{}) *Method {
	return &Method{
		Message:       Message{Type: "method", ID: id},
		ServiceMethod: serviceMethod,
		Args:          args,
	}
}

// Sub is used to send a subscription request to the server.
type Sub struct {
	Message
	SubName string        `json:"name"`
	Args    []interface{} `json:"params"`
}

// NewSub creates a new sub object.
func NewSub(id, subName string, args []interface{}) *Sub {
	return &Sub{
		Message: Message{Type: "sub", ID: id},
		SubName: subName,
		Args:    args,
	}
}
