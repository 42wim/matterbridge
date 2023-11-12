package protocol

func (m *Messenger) StartDiscV5() error {
	return m.transport.StartDiscV5()
}

func (m *Messenger) StopDiscV5() error {
	return m.transport.StopDiscV5()
}
