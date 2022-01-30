package protocol

type CiphertextMessage interface {
	Serialize() []byte
	Type() uint32
}

type GroupCiphertextMessage interface {
	CiphertextMessage
	SignedSerialize() []byte
}

const UnsupportedVersion = 1
const CurrentVersion = 3

const WHISPER_TYPE = 2
const PREKEY_TYPE = 3
const SENDERKEY_TYPE = 4
const SENDERKEY_DISTRIBUTION_TYPE = 5
