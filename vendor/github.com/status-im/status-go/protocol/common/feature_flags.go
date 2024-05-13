package common

type FeatureFlags struct {
	// Datasync indicates whether direct messages should be sent exclusively
	// using datasync, breaking change for non-v1 clients. Public messages
	// are not impacted
	Datasync bool

	// PushNotification indicates whether we should be enabling the push notification feature
	PushNotifications bool

	// MailserverCycle indicates whether we should enable or not the mailserver cycle
	MailserverCycle bool

	// DisableCheckingForBackup disables backup loop
	DisableCheckingForBackup bool

	// DisableAutoMessageLoop disables auto message loop
	DisableAutoMessageLoop bool

	// ResendRawMessagesDisabled indicates whether we should be disabling sending raw messages
	ResendRawMessagesDisabled bool

	// StoreNodesDisabled indicates whether we should fetch messages from store nodes
	StoreNodesDisabled bool

	// Peersyncing indicates whether we should advertise and sync messages with other peers
	Peersyncing bool
}
