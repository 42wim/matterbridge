package peersyncing

type PeerSyncing struct {
	persistence SyncMessagePersistence
	config      Config
}

func New(config Config) *PeerSyncing {
	syncMessagePersistence := config.SyncMessagePersistence
	if syncMessagePersistence == nil {
		syncMessagePersistence = NewSyncMessageSQLitePersistence(config.Database)
	}

	return &PeerSyncing{
		config:      config,
		persistence: syncMessagePersistence,
	}
}

func (p *PeerSyncing) Add(message SyncMessage) error {
	return p.persistence.Add(message)
}

func (p *PeerSyncing) AvailableMessages() ([]SyncMessage, error) {
	return p.persistence.All()
}

func (p *PeerSyncing) AvailableMessagesByGroupID(groupID []byte, limit int) ([]SyncMessage, error) {
	return p.persistence.ByGroupID(groupID, limit)
}

func (p *PeerSyncing) AvailableMessagesByGroupIDs(groupIDs [][]byte, limit int) ([]SyncMessage, error) {
	return p.persistence.ByGroupIDs(groupIDs, limit)
}

func (p *PeerSyncing) MessagesByIDs(messageIDs [][]byte) ([]SyncMessage, error) {
	return p.persistence.ByMessageIDs(messageIDs)
}

func (p *PeerSyncing) OnOffer(messages []SyncMessage) ([]SyncMessage, error) {
	return p.persistence.Complement(messages)
}
