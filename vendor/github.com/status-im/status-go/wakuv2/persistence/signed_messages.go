package persistence

import (
	"crypto/ecdsa"
	"database/sql"
	"errors"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/crypto"
)

// DBStore is a MessageProvider that has a *sql.DB connection
type ProtectedTopicsStore struct {
	db  *sql.DB
	log *zap.Logger

	insertStmt       *sql.Stmt
	fetchPrivKeyStmt *sql.Stmt
	deleteStmt       *sql.Stmt
}

// Creates a new DB store using the db specified via options.
// It will create a messages table if it does not exist and
// clean up records according to the retention policy used
func NewProtectedTopicsStore(log *zap.Logger, db *sql.DB) (*ProtectedTopicsStore, error) {
	insertStmt, err := db.Prepare("INSERT OR REPLACE INTO pubsubtopic_signing_key (topic, priv_key, pub_key) VALUES (?, ?, ?)")
	if err != nil {
		return nil, err
	}

	fetchPrivKeyStmt, err := db.Prepare("SELECT priv_key FROM pubsubtopic_signing_key WHERE topic = ?")
	if err != nil {
		return nil, err
	}

	deleteStmt, err := db.Prepare("DELETE FROM pubsubtopic_signing_key WHERE topic = ?")
	if err != nil {
		return nil, err
	}

	result := new(ProtectedTopicsStore)
	result.log = log.Named("protected-topics-store")
	result.db = db
	result.insertStmt = insertStmt
	result.fetchPrivKeyStmt = fetchPrivKeyStmt
	result.deleteStmt = deleteStmt

	return result, nil
}

func (p *ProtectedTopicsStore) Close() error {
	err := p.insertStmt.Close()
	if err != nil {
		return err
	}

	return p.fetchPrivKeyStmt.Close()
}

func (p *ProtectedTopicsStore) Insert(pubsubTopic string, privKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) error {
	var privKeyBytes []byte
	if privKey != nil {
		privKeyBytes = crypto.FromECDSA(privKey)
	}

	pubKeyBytes := crypto.FromECDSAPub(publicKey)

	_, err := p.insertStmt.Exec(pubsubTopic, privKeyBytes, pubKeyBytes)

	return err
}

func (p *ProtectedTopicsStore) Delete(pubsubTopic string) error {
	_, err := p.deleteStmt.Exec(pubsubTopic)
	return err
}

func (p *ProtectedTopicsStore) FetchPrivateKey(topic string) (privKey *ecdsa.PrivateKey, err error) {
	var privKeyBytes []byte
	err = p.fetchPrivKeyStmt.QueryRow(topic).Scan(&privKeyBytes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return crypto.ToECDSA(privKeyBytes)
}

type ProtectedTopic struct {
	PubKey *ecdsa.PublicKey
	Topic  string
}

func (p *ProtectedTopicsStore) ProtectedTopics() ([]ProtectedTopic, error) {
	rows, err := p.db.Query("SELECT pub_key, topic FROM pubsubtopic_signing_key")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ProtectedTopic
	for rows.Next() {
		var pubKeyBytes []byte
		var topic string
		err := rows.Scan(&pubKeyBytes, &topic)
		if err != nil {
			return nil, err
		}

		pubk, err := crypto.UnmarshalPubkey(pubKeyBytes)
		if err != nil {
			return nil, err
		}

		result = append(result, ProtectedTopic{
			PubKey: pubk,
			Topic:  topic,
		})
	}

	return result, nil
}
