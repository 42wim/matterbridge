package protocol

import (
	"github.com/status-im/status-go/protocol/requests"
)

type WalletConnectSession struct {
	PeerID   string `json:"peerId"`
	DAppName string `json:"dappName"`
	DAppURL  string `json:"dappURL"`
	Info     string `json:"info"`
}

func (m *Messenger) getWalletConnectSession() ([]WalletConnectSession, error) {
	return m.persistence.GetWalletConnectSession()
}

func (m *Messenger) AddWalletConnectSession(request *requests.AddWalletConnectSession) error {
	if err := request.Validate(); err != nil {
		return err
	}

	session := &WalletConnectSession{
		PeerID:   request.PeerID,
		DAppName: request.DAppName,
		DAppURL:  request.DAppURL,
		Info:     request.Info,
	}

	return m.persistence.InsertWalletConnectSession(session)
}

func (m *Messenger) GetWalletConnectSession() ([]WalletConnectSession, error) {

	return m.getWalletConnectSession()
}

func (m *Messenger) DestroyWalletConnectSession(peerID string) error {
	return m.persistence.DeleteWalletConnectSession(peerID)
}
