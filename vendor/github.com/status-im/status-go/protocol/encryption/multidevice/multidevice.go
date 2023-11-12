package multidevice

import (
	"crypto/ecdsa"
	"database/sql"

	"github.com/status-im/status-go/eth-node/crypto"
)

type InstallationMetadata struct {
	// The name of the device
	Name string `json:"name"`
	// The type of device
	DeviceType string `json:"deviceType"`
	// The FCMToken for mobile devices
	FCMToken string `json:"fcmToken"`
}

type Installation struct {
	// Identity is the string identity of the owner
	Identity string `json:"identity"`
	// The installation-id of the device
	ID string `json:"id"`
	// The last known protocol version of the device
	Version uint32 `json:"version"`
	// Enabled is whether the installation is enabled
	Enabled bool `json:"enabled"`
	// Timestamp is the last time we saw this device
	Timestamp int64 `json:"timestamp"`
	// InstallationMetadata
	InstallationMetadata *InstallationMetadata `json:"metadata"`
}

type Config struct {
	MaxInstallations int
	ProtocolVersion  uint32
	InstallationID   string
}

type Multidevice struct {
	persistence *sqlitePersistence
	config      *Config
}

func New(db *sql.DB, config *Config) *Multidevice {
	return &Multidevice{
		config:      config,
		persistence: newSQLitePersistence(db),
	}
}

func (s *Multidevice) InstallationID() string {
	return s.config.InstallationID
}

func (s *Multidevice) GetActiveInstallations(identity *ecdsa.PublicKey) ([]*Installation, error) {
	identityC := crypto.CompressPubkey(identity)
	return s.persistence.GetActiveInstallations(s.config.MaxInstallations, identityC)
}

func (s *Multidevice) GetOurActiveInstallations(identity *ecdsa.PublicKey) ([]*Installation, error) {
	identityC := crypto.CompressPubkey(identity)
	installations, err := s.persistence.GetActiveInstallations(s.config.MaxInstallations-1, identityC)
	if err != nil {
		return nil, err
	}

	installations = append(installations, &Installation{
		ID:      s.config.InstallationID,
		Version: s.config.ProtocolVersion,
	})

	return installations, nil
}

func (s *Multidevice) GetOurInstallations(identity *ecdsa.PublicKey) ([]*Installation, error) {
	var found bool
	identityC := crypto.CompressPubkey(identity)
	installations, err := s.persistence.GetInstallations(identityC)
	if err != nil {
		return nil, err
	}

	for _, installation := range installations {
		if installation.ID == s.config.InstallationID {
			found = true
			installation.Enabled = true
			installation.Version = s.config.ProtocolVersion
		}

	}
	if !found {
		installations = append(installations, &Installation{
			ID:      s.config.InstallationID,
			Enabled: true,
			Version: s.config.ProtocolVersion,
		})
	}

	return installations, nil
}

func (s *Multidevice) AddInstallations(identity []byte, timestamp int64, installations []*Installation, defaultEnabled bool) ([]*Installation, error) {
	return s.persistence.AddInstallations(identity, timestamp, installations, defaultEnabled)
}

func (s *Multidevice) SetInstallationMetadata(identity *ecdsa.PublicKey, installationID string, metadata *InstallationMetadata) error {
	identityC := crypto.CompressPubkey(identity)
	return s.persistence.SetInstallationMetadata(identityC, installationID, metadata)
}

func (s *Multidevice) SetInstallationName(identity *ecdsa.PublicKey, installationID string, name string) error {
	identityC := crypto.CompressPubkey(identity)
	return s.persistence.SetInstallationName(identityC, installationID, name)
}

func (s *Multidevice) EnableInstallation(identity *ecdsa.PublicKey, installationID string) error {
	identityC := crypto.CompressPubkey(identity)
	return s.persistence.EnableInstallation(identityC, installationID)
}

func (s *Multidevice) DisableInstallation(myIdentityKey *ecdsa.PublicKey, installationID string) error {
	myIdentityKeyC := crypto.CompressPubkey(myIdentityKey)
	return s.persistence.DisableInstallation(myIdentityKeyC, installationID)
}
