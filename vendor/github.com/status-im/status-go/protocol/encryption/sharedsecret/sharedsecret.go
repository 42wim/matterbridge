package sharedsecret

import (
	"bytes"
	"crypto/ecdsa"
	"database/sql"
	"errors"

	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/crypto/ecies"
)

const sskLen = 16

type Secret struct {
	Identity *ecdsa.PublicKey
	Key      []byte
}

// SharedSecret generates and manages negotiated secrets.
// Identities (public keys) stored by SharedSecret
// are compressed.
// TODO: make compression of public keys a responsibility  of sqlitePersistence instead of SharedSecret.
type SharedSecret struct {
	persistence *sqlitePersistence
	logger      *zap.Logger
}

func New(db *sql.DB, logger *zap.Logger) *SharedSecret {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &SharedSecret{
		persistence: newSQLitePersistence(db),
		logger:      logger.With(zap.Namespace("SharedSecret")),
	}
}

func (s *SharedSecret) generate(myPrivateKey *ecdsa.PrivateKey, theirPublicKey *ecdsa.PublicKey, installationID string) (*Secret, error) {
	sharedKey, err := ecies.ImportECDSA(myPrivateKey).GenerateShared(
		ecies.ImportECDSAPublic(theirPublicKey),
		sskLen,
		sskLen,
	)
	if err != nil {
		return nil, err
	}

	theirIdentity := crypto.CompressPubkey(theirPublicKey)
	if err = s.persistence.Add(theirIdentity, sharedKey, installationID); err != nil {
		return nil, err
	}

	return &Secret{Key: sharedKey, Identity: theirPublicKey}, err
}

// Generate will generate a shared secret for a given identity, and return it.
func (s *SharedSecret) Generate(myPrivateKey *ecdsa.PrivateKey, theirPublicKey *ecdsa.PublicKey, installationID string) (*Secret, error) {
	return s.generate(myPrivateKey, theirPublicKey, installationID)
}

// Agreed returns true if a secret has been acknowledged by all the installationIDs.
func (s *SharedSecret) Agreed(myPrivateKey *ecdsa.PrivateKey, myInstallationID string, theirPublicKey *ecdsa.PublicKey, theirInstallationIDs []string) (*Secret, bool, error) {
	secret, err := s.generate(myPrivateKey, theirPublicKey, myInstallationID)
	if err != nil {
		return nil, false, err
	}

	if len(theirInstallationIDs) == 0 {
		return secret, false, nil
	}

	theirIdentity := crypto.CompressPubkey(theirPublicKey)
	response, err := s.persistence.Get(theirIdentity, theirInstallationIDs)
	if err != nil {
		return nil, false, err
	}

	for _, installationID := range theirInstallationIDs {
		if !response.installationIDs[installationID] {
			return secret, false, nil
		}
	}

	if !bytes.Equal(secret.Key, response.secret) {
		return nil, false, errors.New("computed and saved secrets are different for a given identity")
	}

	return secret, true, nil
}

func (s *SharedSecret) All() ([]*Secret, error) {
	var secrets []*Secret
	tuples, err := s.persistence.All()
	if err != nil {
		return nil, err
	}

	for _, tuple := range tuples {
		key, err := crypto.DecompressPubkey(tuple[0])
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, &Secret{Identity: key, Key: tuple[1]})
	}

	return secrets, nil
}
