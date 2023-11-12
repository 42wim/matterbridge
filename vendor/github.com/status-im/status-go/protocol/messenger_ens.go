package protocol

func (m *Messenger) ENSVerified(pubkey, ensName string) error {
	clock := m.getTimesource().GetCurrentTime()
	return m.ensVerifier.ENSVerified(pubkey, ensName, clock)
}
