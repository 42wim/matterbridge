package doubleratchet

type SessionStorage interface {
	// Save state keyed by id
	Save(id []byte, state *State) error

	// Load state by id
	Load(id []byte) (*State, error)
}
