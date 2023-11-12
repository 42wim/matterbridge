package pushnotificationserver

import (
	"crypto/ecdsa"
	"database/sql"
	"strings"

	"github.com/golang/protobuf/proto"
	sqlite3 "github.com/mutecomm/go-sqlcipher/v4"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/protobuf"
)

type Persistence interface {
	// GetPushNotificationRegistrationByPublicKeyAndInstallationID retrieve a push notification registration from storage given a public key and installation id
	GetPushNotificationRegistrationByPublicKeyAndInstallationID(publicKey []byte, installationID string) (*protobuf.PushNotificationRegistration, error)
	// GetPushNotificationRegistrationByPublicKey retrieve all the push notification registrations from storage given a public key
	GetPushNotificationRegistrationByPublicKeys(publicKeys [][]byte) ([]*PushNotificationIDAndRegistration, error)
	// GetPushNotificationRegistrationPublicKeys return all the public keys stored
	GetPushNotificationRegistrationPublicKeys() ([][]byte, error)

	//GetPushNotificationRegistrationVersion returns the latest version or 0 for a given pk and installationID
	GetPushNotificationRegistrationVersion(publicKey []byte, installationID string) (uint64, error)
	// UnregisterPushNotificationRegistration unregister a given pk/installationID
	UnregisterPushNotificationRegistration(publicKey []byte, installationID string, version uint64) error

	// DeletePushNotificationRegistration deletes a push notification registration from storage given a public key and installation id
	DeletePushNotificationRegistration(publicKey []byte, installationID string) error
	// SavePushNotificationRegistration saves a push notification option to the db
	SavePushNotificationRegistration(publicKey []byte, registration *protobuf.PushNotificationRegistration) error
	// GetIdentity returns the server identity key
	GetIdentity() (*ecdsa.PrivateKey, error)
	// SaveIdentity saves the server identity key
	SaveIdentity(*ecdsa.PrivateKey) error
	// PushNotificationExists checks whether a push notification exists and inserts it otherwise
	PushNotificationExists([]byte) (bool, error)
}

type SQLitePersistence struct {
	db *sql.DB
}

func NewSQLitePersistence(db *sql.DB) Persistence {
	return &SQLitePersistence{db: db}
}

func (p *SQLitePersistence) GetPushNotificationRegistrationByPublicKeyAndInstallationID(publicKey []byte, installationID string) (*protobuf.PushNotificationRegistration, error) {
	var marshaledRegistration []byte
	err := p.db.QueryRow(`SELECT registration FROM push_notification_server_registrations WHERE public_key = ? AND installation_id = ? AND registration IS NOT NULL`, publicKey, installationID).Scan(&marshaledRegistration)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	registration := &protobuf.PushNotificationRegistration{}
	if err := proto.Unmarshal(marshaledRegistration, registration); err != nil {
		return nil, err
	}
	return registration, nil
}

func (p *SQLitePersistence) GetPushNotificationRegistrationVersion(publicKey []byte, installationID string) (uint64, error) {
	var version uint64
	err := p.db.QueryRow(`SELECT version FROM push_notification_server_registrations WHERE public_key = ? AND installation_id = ?`, publicKey, installationID).Scan(&version)

	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return version, nil
}

type PushNotificationIDAndRegistration struct {
	ID           []byte
	Registration *protobuf.PushNotificationRegistration
}

func (p *SQLitePersistence) GetPushNotificationRegistrationByPublicKeys(publicKeys [][]byte) ([]*PushNotificationIDAndRegistration, error) {
	// TODO: check for a max number of keys

	publicKeyArgs := make([]interface{}, 0, len(publicKeys))
	for _, pk := range publicKeys {
		publicKeyArgs = append(publicKeyArgs, pk)
	}

	inVector := strings.Repeat("?, ", len(publicKeys)-1) + "?"

	rows, err := p.db.Query(`SELECT public_key,registration FROM push_notification_server_registrations WHERE registration IS NOT NULL AND public_key IN (`+inVector+`)`, publicKeyArgs...) // nolint: gosec
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var registrations []*PushNotificationIDAndRegistration
	for rows.Next() {
		response := &PushNotificationIDAndRegistration{}
		var marshaledRegistration []byte
		err := rows.Scan(&response.ID, &marshaledRegistration)
		if err != nil {
			return nil, err
		}

		registration := &protobuf.PushNotificationRegistration{}
		// Skip if there's no registration
		if marshaledRegistration == nil {
			continue
		}

		if err := proto.Unmarshal(marshaledRegistration, registration); err != nil {
			return nil, err
		}
		response.Registration = registration
		registrations = append(registrations, response)
	}
	return registrations, nil
}

func (p *SQLitePersistence) GetPushNotificationRegistrationPublicKeys() ([][]byte, error) {
	rows, err := p.db.Query(`SELECT public_key FROM push_notification_server_registrations WHERE registration IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var publicKeys [][]byte
	for rows.Next() {
		var publicKey []byte
		err := rows.Scan(&publicKey)
		if err != nil {
			return nil, err
		}

		publicKeys = append(publicKeys, publicKey)
	}
	return publicKeys, nil
}

func (p *SQLitePersistence) SavePushNotificationRegistration(publicKey []byte, registration *protobuf.PushNotificationRegistration) error {
	marshaledRegistration, err := proto.Marshal(registration)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(`INSERT INTO push_notification_server_registrations (public_key, installation_id, version, registration) VALUES (?, ?, ?, ?)`, publicKey, registration.InstallationId, registration.Version, marshaledRegistration)
	return err
}

func (p *SQLitePersistence) UnregisterPushNotificationRegistration(publicKey []byte, installationID string, version uint64) error {
	_, err := p.db.Exec(`UPDATE push_notification_server_registrations SET registration = NULL, version = ? WHERE public_key = ? AND installation_id = ?`, version, publicKey, installationID)
	return err
}

func (p *SQLitePersistence) DeletePushNotificationRegistration(publicKey []byte, installationID string) error {
	_, err := p.db.Exec(`DELETE FROM push_notification_server_registrations WHERE public_key = ? AND installation_id = ?`, publicKey, installationID)
	return err
}

func (p *SQLitePersistence) SaveIdentity(privateKey *ecdsa.PrivateKey) error {
	_, err := p.db.Exec(`INSERT INTO push_notification_server_identity (private_key) VALUES (?)`, crypto.FromECDSA(privateKey))
	return err
}

func (p *SQLitePersistence) GetIdentity() (*ecdsa.PrivateKey, error) {
	var pkBytes []byte
	err := p.db.QueryRow(`SELECT private_key FROM push_notification_server_identity LIMIT 1`).Scan(&pkBytes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	pk, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

func (p *SQLitePersistence) PushNotificationExists(messageID []byte) (bool, error) {
	_, err := p.db.Exec(`INSERT INTO push_notification_server_notifications  VALUES (?)`, messageID)
	if err != nil && err.(sqlite3.Error).ExtendedCode == sqlite3.ErrConstraintUnique {
		return true, nil
	} else if err != nil {
		return false, err
	}

	return false, nil
}
