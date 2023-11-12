package signal

const (
	// ReEncryptionStarted is sent when db reencryption was started.
	ReEncryptionStarted = "db.reEncryption.started"
	// ReEncryptionFinished is sent when db reencryption was finished.
	ReEncryptionFinished = "db.reEncryption.finished"
)

// Send db.reencryption.started signal.
func SendReEncryptionStarted() {
	send(ReEncryptionStarted, nil)
}

// Send db.reencryption.finished signal.
func SendReEncryptionFinished() {
	send(ReEncryptionFinished, nil)
}
