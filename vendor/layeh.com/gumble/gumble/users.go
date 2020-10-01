package gumble

// Users is a map of server users.
//
// When accessed through client.Users, it contains all users currently on the
// server. When accessed through a specific channel
// (e.g. client.Channels[0].Users), it contains only the users in the
// channel.
type Users map[uint32]*User

// create adds a new user with the given session to the collection. If a user
// with the given session already exists, it is overwritten.
func (u Users) create(session uint32) *User {
	user := &User{
		Session: session,
	}
	u[session] = user
	return user
}

// Find returns the user with the given name. nil is returned if no user exists
// with the given name.
func (u Users) Find(name string) *User {
	for _, user := range u {
		if user.Name == name {
			return user
		}
	}
	return nil
}
