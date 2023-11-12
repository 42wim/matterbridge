package encryption

import (
	"crypto/ecdsa"
	"database/sql"
	"errors"
	"strings"

	dr "github.com/status-im/doubleratchet"

	"github.com/status-im/status-go/eth-node/crypto"

	"github.com/status-im/status-go/protocol/encryption/multidevice"
)

// RatchetInfo holds the current ratchet state.
type RatchetInfo struct {
	ID             []byte
	Sk             []byte
	PrivateKey     []byte
	PublicKey      []byte
	Identity       []byte
	BundleID       []byte
	EphemeralKey   []byte
	InstallationID string
}

// A safe max number of rows.
const maxNumberOfRows = 100000000

type sqlitePersistence struct {
	DB             *sql.DB
	keysStorage    dr.KeysStorage
	sessionStorage dr.SessionStorage
}

func newSQLitePersistence(db *sql.DB) *sqlitePersistence {
	return &sqlitePersistence{
		DB:             db,
		keysStorage:    newSQLiteKeysStorage(db),
		sessionStorage: newSQLiteSessionStorage(db),
	}
}

// GetKeysStorage returns the associated double ratchet KeysStorage object
func (s *sqlitePersistence) KeysStorage() dr.KeysStorage {
	return s.keysStorage
}

// GetSessionStorage returns the associated double ratchet SessionStorage object
func (s *sqlitePersistence) SessionStorage() dr.SessionStorage {
	return s.sessionStorage
}

// AddPrivateBundle adds the specified BundleContainer to the database
func (s *sqlitePersistence) AddPrivateBundle(bc *BundleContainer) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	for installationID, signedPreKey := range bc.GetBundle().GetSignedPreKeys() {
		var version uint32
		stmt, err := tx.Prepare(`SELECT version
					 FROM bundles
					 WHERE installation_id = ? AND identity = ?
					 ORDER BY version DESC
					 LIMIT 1`)
		if err != nil {
			return err
		}

		defer stmt.Close()

		err = stmt.QueryRow(installationID, bc.GetBundle().GetIdentity()).Scan(&version)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		stmt, err = tx.Prepare(`INSERT INTO bundles(identity, private_key, signed_pre_key, installation_id, version, timestamp)
					VALUES(?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(
			bc.GetBundle().GetIdentity(),
			bc.GetPrivateSignedPreKey(),
			signedPreKey.GetSignedPreKey(),
			installationID,
			version+1,
			bc.GetBundle().GetTimestamp(),
		)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

// AddPublicBundle adds the specified Bundle to the database
func (s *sqlitePersistence) AddPublicBundle(b *Bundle) error {
	tx, err := s.DB.Begin()

	if err != nil {
		return err
	}

	for installationID, signedPreKeyContainer := range b.GetSignedPreKeys() {
		signedPreKey := signedPreKeyContainer.GetSignedPreKey()
		version := signedPreKeyContainer.GetVersion()
		insertStmt, err := tx.Prepare(`INSERT INTO bundles(identity, signed_pre_key, installation_id, version, timestamp)
					       VALUES( ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer insertStmt.Close()

		_, err = insertStmt.Exec(
			b.GetIdentity(),
			signedPreKey,
			installationID,
			version,
			b.GetTimestamp(),
		)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		// Mark old bundles as expired
		updateStmt, err := tx.Prepare(`UPDATE bundles
					       SET expired = 1
					       WHERE identity = ? AND installation_id = ? AND version < ?`)
		if err != nil {
			return err
		}
		defer updateStmt.Close()

		_, err = updateStmt.Exec(
			b.GetIdentity(),
			installationID,
			version,
		)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

	}

	return tx.Commit()
}

// GetAnyPrivateBundle retrieves any bundle from the database containing a private key
func (s *sqlitePersistence) GetAnyPrivateBundle(myIdentityKey []byte, installations []*multidevice.Installation) (*BundleContainer, error) {

	versions := make(map[string]uint32)
	/* #nosec */
	statement := `SELECT identity, private_key, signed_pre_key, installation_id, timestamp, version
	              FROM bundles
		      WHERE expired = 0 AND identity = ? AND installation_id IN (?` + strings.Repeat(",?", len(installations)-1) + ")"
	stmt, err := s.DB.Prepare(statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var timestamp int64
	var identity []byte
	var privateKey []byte
	var version uint32

	args := make([]interface{}, len(installations)+1)
	args[0] = myIdentityKey
	for i, installation := range installations {
		// Lookup up map for versions
		versions[installation.ID] = installation.Version

		args[i+1] = installation.ID
	}

	rows, err := stmt.Query(args...)
	rowCount := 0

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	bundle := &Bundle{
		SignedPreKeys: make(map[string]*SignedPreKey),
	}

	bundleContainer := &BundleContainer{
		Bundle: bundle,
	}

	for rows.Next() {
		var signedPreKey []byte
		var installationID string
		rowCount++
		err = rows.Scan(
			&identity,
			&privateKey,
			&signedPreKey,
			&installationID,
			&timestamp,
			&version,
		)
		if err != nil {
			return nil, err
		}
		// If there is a private key, we set the timestamp of the bundle container
		if privateKey != nil {
			bundle.Timestamp = timestamp
		}

		bundle.SignedPreKeys[installationID] = &SignedPreKey{
			SignedPreKey:    signedPreKey,
			Version:         version,
			ProtocolVersion: versions[installationID],
		}
		bundle.Identity = identity
	}

	// If no records are found or no record with private key, return nil
	if rowCount == 0 || bundleContainer.GetBundle().Timestamp == 0 {
		return nil, nil
	}

	return bundleContainer, nil

}

// GetPrivateKeyBundle retrieves a private key for a bundle from the database
func (s *sqlitePersistence) GetPrivateKeyBundle(bundleID []byte) ([]byte, error) {
	stmt, err := s.DB.Prepare(`SELECT private_key
				   FROM bundles
				   WHERE signed_pre_key = ? LIMIT 1`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var privateKey []byte

	err = stmt.QueryRow(bundleID).Scan(&privateKey)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return privateKey, nil
	default:
		return nil, err
	}
}

// MarkBundleExpired expires any private bundle for a given identity
func (s *sqlitePersistence) MarkBundleExpired(identity []byte) error {
	stmt, err := s.DB.Prepare(`UPDATE bundles
				   SET expired = 1
				   WHERE identity = ? AND private_key IS NOT NULL`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(identity)

	return err
}

// GetPublicBundle retrieves an existing Bundle for the specified public key from the database
func (s *sqlitePersistence) GetPublicBundle(publicKey *ecdsa.PublicKey, installations []*multidevice.Installation) (*Bundle, error) {

	if len(installations) == 0 {
		return nil, nil
	}

	versions := make(map[string]uint32)
	identity := crypto.CompressPubkey(publicKey)

	/* #nosec */
	statement := `SELECT signed_pre_key,installation_id, version
		      FROM bundles
		      WHERE expired = 0 AND identity = ? AND installation_id IN (?` + strings.Repeat(",?", len(installations)-1) + `)
		      ORDER BY version DESC`
	stmt, err := s.DB.Prepare(statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	args := make([]interface{}, len(installations)+1)
	args[0] = identity
	for i, installation := range installations {
		// Lookup up map for versions
		versions[installation.ID] = installation.Version
		args[i+1] = installation.ID
	}

	rows, err := stmt.Query(args...)
	rowCount := 0

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	bundle := &Bundle{
		Identity:      identity,
		SignedPreKeys: make(map[string]*SignedPreKey),
	}

	for rows.Next() {
		var signedPreKey []byte
		var installationID string
		var version uint32
		rowCount++
		err = rows.Scan(
			&signedPreKey,
			&installationID,
			&version,
		)
		if err != nil {
			return nil, err
		}

		bundle.SignedPreKeys[installationID] = &SignedPreKey{
			SignedPreKey:    signedPreKey,
			Version:         version,
			ProtocolVersion: versions[installationID],
		}

	}

	if rowCount == 0 {
		return nil, nil
	}

	return bundle, nil

}

// AddRatchetInfo persists the specified ratchet info into the database
func (s *sqlitePersistence) AddRatchetInfo(key []byte, identity []byte, bundleID []byte, ephemeralKey []byte, installationID string) error {
	stmt, err := s.DB.Prepare(`INSERT INTO ratchet_info_v2(symmetric_key, identity, bundle_id, ephemeral_key, installation_id)
				   VALUES(?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		key,
		identity,
		bundleID,
		ephemeralKey,
		installationID,
	)

	return err
}

// GetRatchetInfo retrieves the existing RatchetInfo for a specified bundle ID and interlocutor public key from the database
func (s *sqlitePersistence) GetRatchetInfo(bundleID []byte, theirIdentity []byte, installationID string) (*RatchetInfo, error) {
	stmt, err := s.DB.Prepare(`SELECT ratchet_info_v2.identity, ratchet_info_v2.symmetric_key, bundles.private_key, bundles.signed_pre_key, ratchet_info_v2.ephemeral_key, ratchet_info_v2.installation_id
				   FROM ratchet_info_v2 JOIN bundles ON bundle_id = signed_pre_key
				   WHERE ratchet_info_v2.identity = ? AND ratchet_info_v2.installation_id = ? AND bundle_id = ?
				   LIMIT 1`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ratchetInfo := &RatchetInfo{
		BundleID: bundleID,
	}

	err = stmt.QueryRow(theirIdentity, installationID, bundleID).Scan(
		&ratchetInfo.Identity,
		&ratchetInfo.Sk,
		&ratchetInfo.PrivateKey,
		&ratchetInfo.PublicKey,
		&ratchetInfo.EphemeralKey,
		&ratchetInfo.InstallationID,
	)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		ratchetInfo.ID = append(bundleID, []byte(ratchetInfo.InstallationID)...)
		return ratchetInfo, nil
	default:
		return nil, err
	}
}

// GetAnyRatchetInfo retrieves any existing RatchetInfo for a specified interlocutor public key from the database
func (s *sqlitePersistence) GetAnyRatchetInfo(identity []byte, installationID string) (*RatchetInfo, error) {
	stmt, err := s.DB.Prepare(`SELECT symmetric_key, bundles.private_key, signed_pre_key, bundle_id, ephemeral_key
				   FROM ratchet_info_v2 JOIN bundles ON bundle_id = signed_pre_key
				   WHERE expired = 0 AND ratchet_info_v2.identity = ? AND ratchet_info_v2.installation_id = ?
				   LIMIT 1`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ratchetInfo := &RatchetInfo{
		Identity:       identity,
		InstallationID: installationID,
	}

	err = stmt.QueryRow(identity, installationID).Scan(
		&ratchetInfo.Sk,
		&ratchetInfo.PrivateKey,
		&ratchetInfo.PublicKey,
		&ratchetInfo.BundleID,
		&ratchetInfo.EphemeralKey,
	)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		ratchetInfo.ID = append(ratchetInfo.BundleID, []byte(installationID)...)
		return ratchetInfo, nil
	default:
		return nil, err
	}
}

// RatchetInfoConfirmed clears the ephemeral key in the RatchetInfo
// associated with the specified bundle ID and interlocutor identity public key
func (s *sqlitePersistence) RatchetInfoConfirmed(bundleID []byte, theirIdentity []byte, installationID string) error {
	stmt, err := s.DB.Prepare(`UPDATE ratchet_info_v2
	                           SET ephemeral_key = NULL
				   WHERE identity = ? AND bundle_id = ? AND installation_id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		theirIdentity,
		bundleID,
		installationID,
	)

	return err
}

type sqliteKeysStorage struct {
	db *sql.DB
}

func newSQLiteKeysStorage(db *sql.DB) *sqliteKeysStorage {
	return &sqliteKeysStorage{
		db: db,
	}
}

// Get retrieves the message key for a specified public key and message number
func (s *sqliteKeysStorage) Get(pubKey dr.Key, msgNum uint) (dr.Key, bool, error) {
	var key []byte
	stmt, err := s.db.Prepare(`SELECT message_key
	                           FROM keys
				   WHERE public_key = ? AND msg_num = ?
				   LIMIT 1`)

	if err != nil {
		return key, false, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(pubKey, msgNum).Scan(&key)
	switch err {
	case sql.ErrNoRows:
		return key, false, nil
	case nil:
		return key, true, nil
	default:
		return key, false, err
	}
}

// Put stores a key with the specified public key, message number and message key
func (s *sqliteKeysStorage) Put(sessionID []byte, pubKey dr.Key, msgNum uint, mk dr.Key, seqNum uint) error {
	stmt, err := s.db.Prepare(`INSERT INTO keys(session_id, public_key, msg_num, message_key, seq_num)
	                           VALUES(?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		sessionID,
		pubKey,
		msgNum,
		mk,
		seqNum,
	)

	return err
}

// DeleteOldMks caps remove any key < seq_num, included
func (s *sqliteKeysStorage) DeleteOldMks(sessionID []byte, deleteUntil uint) error {
	stmt, err := s.db.Prepare(`DELETE FROM keys
	                           WHERE session_id = ? AND seq_num <= ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		sessionID,
		deleteUntil,
	)

	return err
}

// TruncateMks caps the number of keys to maxKeysPerSession deleting them in FIFO fashion
func (s *sqliteKeysStorage) TruncateMks(sessionID []byte, maxKeysPerSession int) error {
	stmt, err := s.db.Prepare(`DELETE FROM keys
				   WHERE rowid IN (SELECT rowid FROM keys WHERE session_id = ? ORDER BY seq_num DESC LIMIT ? OFFSET ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		sessionID,
		// We LIMIT to the max number of rows here, as OFFSET can't be used without a LIMIT
		maxNumberOfRows,
		maxKeysPerSession,
	)

	return err
}

// DeleteMk deletes the key with the specified public key and message key
func (s *sqliteKeysStorage) DeleteMk(pubKey dr.Key, msgNum uint) error {
	stmt, err := s.db.Prepare(`DELETE FROM keys
				   WHERE public_key = ? AND msg_num = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		pubKey,
		msgNum,
	)

	return err
}

// Count returns the count of keys with the specified public key
func (s *sqliteKeysStorage) Count(pubKey dr.Key) (uint, error) {
	stmt, err := s.db.Prepare(`SELECT COUNT(1)
				   FROM keys
				   WHERE public_key = ?`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var count uint
	err = stmt.QueryRow(pubKey).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CountAll returns the count of keys with the specified public key
func (s *sqliteKeysStorage) CountAll() (uint, error) {
	stmt, err := s.db.Prepare(`SELECT COUNT(1)
				   FROM keys`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var count uint
	err = stmt.QueryRow().Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// All returns nil
func (s *sqliteKeysStorage) All() (map[string]map[uint]dr.Key, error) {
	return nil, nil
}

type sqliteSessionStorage struct {
	db *sql.DB
}

func newSQLiteSessionStorage(db *sql.DB) *sqliteSessionStorage {
	return &sqliteSessionStorage{
		db: db,
	}
}

// Save persists the specified double ratchet state
func (s *sqliteSessionStorage) Save(id []byte, state *dr.State) error {
	dhr := state.DHr
	dhs := state.DHs
	dhsPublic := dhs.PublicKey()
	dhsPrivate := dhs.PrivateKey()
	pn := state.PN
	step := state.Step
	keysCount := state.KeysCount

	rootChainKey := state.RootCh.CK

	sendChainKey := state.SendCh.CK
	sendChainN := state.SendCh.N

	recvChainKey := state.RecvCh.CK
	recvChainN := state.RecvCh.N

	stmt, err := s.db.Prepare(`INSERT INTO sessions(id, dhr, dhs_public, dhs_private, root_chain_key, send_chain_key, send_chain_n, recv_chain_key, recv_chain_n, pn, step, keys_count)
				   VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		id,
		dhr,
		dhsPublic,
		dhsPrivate,
		rootChainKey,
		sendChainKey,
		sendChainN,
		recvChainKey,
		recvChainN,
		pn,
		step,
		keysCount,
	)

	return err
}

// Load retrieves the double ratchet state for a given ID
func (s *sqliteSessionStorage) Load(id []byte) (*dr.State, error) {
	stmt, err := s.db.Prepare(`SELECT dhr, dhs_public, dhs_private, root_chain_key, send_chain_key, send_chain_n, recv_chain_key, recv_chain_n, pn, step, keys_count
				   FROM sessions
				   WHERE id = ?`)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var (
		dhr          []byte
		dhsPublic    []byte
		dhsPrivate   []byte
		rootChainKey []byte
		sendChainKey []byte
		sendChainN   uint
		recvChainKey []byte
		recvChainN   uint
		pn           uint
		step         uint
		keysCount    uint
	)

	err = stmt.QueryRow(id).Scan(
		&dhr,
		&dhsPublic,
		&dhsPrivate,
		&rootChainKey,
		&sendChainKey,
		&sendChainN,
		&recvChainKey,
		&recvChainN,
		&pn,
		&step,
		&keysCount,
	)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		state := dr.DefaultState(rootChainKey)

		state.PN = uint32(pn)
		state.Step = step
		state.KeysCount = keysCount

		state.DHs = crypto.DHPair{
			PrvKey: dhsPrivate,
			PubKey: dhsPublic,
		}

		state.DHr = dhr

		state.SendCh.CK = sendChainKey
		state.SendCh.N = uint32(sendChainN)

		state.RecvCh.CK = recvChainKey
		state.RecvCh.N = uint32(recvChainN)

		return &state, nil
	default:
		return nil, err
	}
}

type HRCache struct {
	GroupID         []byte
	KeyID           []byte
	DeprecatedKeyID uint32
	Key             []byte
	Hash            []byte
	SeqNo           uint32
}

// GetHashRatchetCache retrieves a hash ratchet key by group ID and seqNo.
// If cache data with given seqNo (e.g. 0) is not found,
// then the query will return the cache data with the latest seqNo
func (s *sqlitePersistence) GetHashRatchetCache(ratchet *HashRatchetKeyCompatibility, seqNo uint32) (*HRCache, error) {
	stmt, err := s.DB.Prepare(`WITH input AS (
       select ? AS group_id, ? AS key_id, ? as seq_no, ? AS old_key_id
     ),
     cec AS (
       SELECT e.key, c.seq_no, c.hash FROM hash_ratchet_encryption e, input i
			 LEFT JOIN hash_ratchet_encryption_cache c ON e.group_id=c.group_id AND (e.key_id=c.key_id OR e.deprecated_key_id=c.key_id)
       WHERE (e.key_id=i.key_id OR e.deprecated_key_id=i.old_key_id) AND e.group_id=i.group_id),
    seq_nos AS (
    select CASE
		  	WHEN EXISTS (SELECT c.seq_no from cec c, input i where c.seq_no=i.seq_no)
				THEN i.seq_no
			  ELSE (select max(seq_no) from cec)
			END as seq_no from input i
    )
		 SELECT c.key, c.seq_no, c.hash FROM cec c, input i, seq_nos s
    where case when not exists(select seq_no from seq_nos where seq_no is not null)
    then 1 else c.seq_no = s.seq_no end`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var key, hash []byte
	var seqNoPtr *uint32

	oldFormat := ratchet.IsOldFormat()
	if oldFormat {
		// Query using the deprecated format
		err = stmt.QueryRow(ratchet.GroupID, nil, seqNo, ratchet.DeprecatedKeyID()).Scan(&key, &seqNoPtr, &hash) //nolint: ineffassign

	} else {
		keyID, err := ratchet.GetKeyID()
		if err != nil {
			return nil, err
		}

		err = stmt.QueryRow(ratchet.GroupID, keyID, seqNo, ratchet.DeprecatedKeyID()).Scan(&key, &seqNoPtr, &hash) //nolint: ineffassign,staticcheck
	}
	if len(hash) == 0 && len(key) == 0 {
		return nil, nil
	}

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		var seqNoResult uint32
		if seqNoPtr == nil {
			seqNoResult = 0
		} else {
			seqNoResult = *seqNoPtr
		}

		ratchet.Key = key
		keyID, err := ratchet.GetKeyID()

		if err != nil {
			return nil, err
		}

		res := &HRCache{
			KeyID: keyID,
			Key:   key,
			Hash:  hash,
			SeqNo: seqNoResult,
		}

		return res, nil
	default:
		return nil, err
	}
}

type HashRatchetKeyCompatibility struct {
	GroupID   []byte
	keyID     []byte
	Timestamp uint64
	Key       []byte
}

func (h *HashRatchetKeyCompatibility) DeprecatedKeyID() uint32 {
	return uint32(h.Timestamp)
}

func (h *HashRatchetKeyCompatibility) IsOldFormat() bool {
	return len(h.keyID) == 0 && len(h.Key) == 0
}

func (h *HashRatchetKeyCompatibility) GetKeyID() ([]byte, error) {
	if len(h.keyID) != 0 {
		return h.keyID, nil
	}

	if len(h.GroupID) == 0 || h.Timestamp == 0 || len(h.Key) == 0 {
		return nil, errors.New("could not create key")
	}

	return generateHashRatchetKeyID(h.GroupID, h.Timestamp, h.Key), nil
}

func (h *HashRatchetKeyCompatibility) GenerateNext() (*HashRatchetKeyCompatibility, error) {

	ratchet := &HashRatchetKeyCompatibility{
		GroupID: h.GroupID,
	}

	// Randomly generate a hash ratchet key
	hrKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	hrKeyBytes := crypto.FromECDSA(hrKey)

	if err != nil {
		return nil, err
	}

	currentTime := GetCurrentTime()
	if h.Timestamp < currentTime {
		ratchet.Timestamp = bumpKeyID(currentTime)
	} else {
		ratchet.Timestamp = h.Timestamp + 1
	}

	ratchet.Key = hrKeyBytes

	_, err = ratchet.GetKeyID()
	if err != nil {
		return nil, err
	}

	return ratchet, nil
}

// GetCurrentKeyForGroup retrieves a key ID for given group ID
// (with an assumption that key ids are shared in the group, and
// at any given time there is a single key used)
func (s *sqlitePersistence) GetCurrentKeyForGroup(groupID []byte) (*HashRatchetKeyCompatibility, error) {
	ratchet := &HashRatchetKeyCompatibility{
		GroupID: groupID,
	}

	stmt, err := s.DB.Prepare(`SELECT key_id, key_timestamp, key
				   FROM hash_ratchet_encryption
				     WHERE group_id = ? order by key_timestamp desc limit 1`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var keyID, key []byte
	var timestamp uint64
	err = stmt.QueryRow(groupID).Scan(&keyID, &timestamp, &key)

	switch err {
	case sql.ErrNoRows:
		return ratchet, nil
	case nil:
		ratchet.Key = key
		ratchet.Timestamp = timestamp
		_, err = ratchet.GetKeyID()
		if err != nil {
			return nil, err
		}
		return ratchet, nil
	default:
		return nil, err
	}
}

// GetKeysForGroup retrieves all key IDs for given group ID
func (s *sqlitePersistence) GetKeysForGroup(groupID []byte) ([]*HashRatchetKeyCompatibility, error) {

	var ratchets []*HashRatchetKeyCompatibility
	stmt, err := s.DB.Prepare(`SELECT key_id, key_timestamp, key
				   FROM hash_ratchet_encryption
				     WHERE group_id = ? order by key_timestamp desc`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(groupID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		ratchet := &HashRatchetKeyCompatibility{GroupID: groupID}
		err := rows.Scan(&ratchet.keyID, &ratchet.Timestamp, &ratchet.Key)
		if err != nil {
			return nil, err
		}
		ratchets = append(ratchets, ratchet)
	}

	return ratchets, nil
}

// SaveHashRatchetKeyHash saves a hash ratchet key cache data
func (s *sqlitePersistence) SaveHashRatchetKeyHash(
	ratchet *HashRatchetKeyCompatibility,
	hash []byte,
	seqNo uint32,
) error {

	stmt, err := s.DB.Prepare(`INSERT INTO hash_ratchet_encryption_cache(group_id, key_id, hash, seq_no)
           VALUES(?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	keyID, err := ratchet.GetKeyID()
	if err != nil {
		return err
	}

	_, err = stmt.Exec(ratchet.GroupID, keyID, hash, seqNo)

	return err
}

// SaveHashRatchetKey saves a hash ratchet key
func (s *sqlitePersistence) SaveHashRatchetKey(ratchet *HashRatchetKeyCompatibility) error {
	stmt, err := s.DB.Prepare(`INSERT INTO hash_ratchet_encryption(group_id, key_id, key_timestamp, deprecated_key_id, key)
           VALUES(?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	keyID, err := ratchet.GetKeyID()
	if err != nil {
		return err
	}

	_, err = stmt.Exec(ratchet.GroupID, keyID, ratchet.Timestamp, ratchet.DeprecatedKeyID(), ratchet.Key)

	return err
}

func (s *sqlitePersistence) GetHashRatchetKeyByID(keyID []byte) (*HashRatchetKeyCompatibility, error) {
	ratchet := &HashRatchetKeyCompatibility{
		keyID: keyID,
	}

	err := s.DB.QueryRow(`
		SELECT group_id, key_timestamp, key
		FROM hash_ratchet_encryption
		WHERE key_id = ?`, keyID).Scan(&ratchet.GroupID, &ratchet.Timestamp, &ratchet.Key)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return ratchet, nil
}
