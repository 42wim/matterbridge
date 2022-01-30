package protocol

// NewSenderKeyName returns a new SenderKeyName object.
func NewSenderKeyName(groupID string, sender *SignalAddress) *SenderKeyName {
	return &SenderKeyName{
		groupID: groupID,
		sender:  sender,
	}
}

// SenderKeyName is a structure for a group session address.
type SenderKeyName struct {
	groupID string
	sender  *SignalAddress
}

// GroupID returns the sender key group id
func (n *SenderKeyName) GroupID() string {
	return n.groupID
}

// Sender returns the Signal address of sending user in the group.
func (n *SenderKeyName) Sender() *SignalAddress {
	return n.sender
}
