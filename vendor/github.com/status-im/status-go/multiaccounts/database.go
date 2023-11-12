package multiaccounts

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/common/dbsetup"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/multiaccounts/common"
	"github.com/status-im/status-go/multiaccounts/migrations"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/sqlite"
)

type ColorHash [][2]int

// Account stores public information about account.
type Account struct {
	Name                    string                    `json:"name"`
	Timestamp               int64                     `json:"timestamp"`
	Identicon               string                    `json:"identicon"`
	ColorHash               ColorHash                 `json:"colorHash"`
	ColorID                 int64                     `json:"colorId"`
	CustomizationColor      common.CustomizationColor `json:"customizationColor,omitempty"`
	KeycardPairing          string                    `json:"keycard-pairing"`
	KeyUID                  string                    `json:"key-uid"`
	Images                  []images.IdentityImage    `json:"images"`
	KDFIterations           int                       `json:"kdfIterations,omitempty"`
	CustomizationColorClock uint64                    `json:"-"`
}

func (a *Account) ToProtobuf() *protobuf.MultiAccount {
	var colorHashes []*protobuf.MultiAccount_ColorHash
	for _, index := range a.ColorHash {
		var i []int64
		for _, is := range index {
			i = append(i, int64(is))
		}

		colorHashes = append(colorHashes, &protobuf.MultiAccount_ColorHash{Index: i})
	}

	var identityImages []*protobuf.MultiAccount_IdentityImage
	for _, ii := range a.Images {
		identityImages = append(identityImages, ii.ToProtobuf())
	}

	return &protobuf.MultiAccount{
		Name:                    a.Name,
		Timestamp:               a.Timestamp,
		Identicon:               a.Identicon,
		ColorHash:               colorHashes,
		ColorId:                 a.ColorID,
		CustomizationColor:      string(a.CustomizationColor),
		KeycardPairing:          a.KeycardPairing,
		KeyUid:                  a.KeyUID,
		Images:                  identityImages,
		CustomizationColorClock: a.CustomizationColorClock,
	}
}

func (a *Account) FromProtobuf(ma *protobuf.MultiAccount) {
	var colorHash ColorHash
	for _, index := range ma.ColorHash {
		var i [2]int
		for n, is := range index.Index {
			i[n] = int(is)
		}

		colorHash = append(colorHash, i)
	}

	var identityImages []images.IdentityImage
	for _, ii := range ma.Images {
		iii := images.IdentityImage{}
		iii.FromProtobuf(ii)
		identityImages = append(identityImages, iii)
	}

	a.Name = ma.Name
	a.Timestamp = ma.Timestamp
	a.Identicon = ma.Identicon
	a.ColorHash = colorHash
	a.ColorID = ma.ColorId
	a.KeycardPairing = ma.KeycardPairing
	a.CustomizationColor = common.CustomizationColor(ma.CustomizationColor)
	a.KeyUID = ma.KeyUid
	a.Images = identityImages
	a.CustomizationColorClock = ma.CustomizationColorClock
}

type MultiAccountMarshaller interface {
	ToMultiAccount() *Account
}

type IdentityImageSubscriptionChange struct {
	PublishExpected bool
}

type Database struct {
	db                         *sql.DB
	identityImageSubscriptions []chan *IdentityImageSubscriptionChange
}

// InitializeDB creates db file at a given path and applies migrations.
func InitializeDB(path string) (*Database, error) {
	db, err := sqlite.OpenUnecryptedDB(path)
	if err != nil {
		return nil, err
	}
	err = migrations.Migrate(db, nil)
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

func (db *Database) Close() error {
	return db.db.Close()
}

func (db *Database) GetAccountKDFIterationsNumber(keyUID string) (kdfIterationsNumber int, err error) {
	err = db.db.QueryRow("SELECT  kdfIterations FROM accounts WHERE keyUid = ?", keyUID).Scan(&kdfIterationsNumber)
	if err != nil {
		return -1, err
	}
	return
}

func (db *Database) GetAccounts() (rst []Account, err error) {
	rows, err := db.db.Query("SELECT  a.name, a.loginTimestamp, a.identicon, a.colorHash, a.colorId, a.customizationColor, a.customizationColorClock, a.keycardPairing, a.keyUid, a.kdfIterations, ii.name, ii.image_payload, ii.width, ii.height, ii.file_size, ii.resize_target, ii.clock FROM accounts AS a LEFT JOIN identity_images AS ii ON ii.key_uid = a.keyUid ORDER BY loginTimestamp DESC")
	if err != nil {
		return nil, err
	}
	defer func() {
		errClose := rows.Close()
		err = valueOr(err, errClose)
	}()

	for rows.Next() {
		acc := Account{}
		accLoginTimestamp := sql.NullInt64{}
		accIdenticon := sql.NullString{}
		accColorHash := sql.NullString{}
		accColorID := sql.NullInt64{}
		ii := &images.IdentityImage{}
		iiName := sql.NullString{}
		iiWidth := sql.NullInt64{}
		iiHeight := sql.NullInt64{}
		iiFileSize := sql.NullInt64{}
		iiResizeTarget := sql.NullInt64{}
		iiClock := sql.NullInt64{}

		err = rows.Scan(
			&acc.Name,
			&accLoginTimestamp,
			&accIdenticon,
			&accColorHash,
			&accColorID,
			&acc.CustomizationColor,
			&acc.CustomizationColorClock,
			&acc.KeycardPairing,
			&acc.KeyUID,
			&acc.KDFIterations,
			&iiName,
			&ii.Payload,
			&iiWidth,
			&iiHeight,
			&iiFileSize,
			&iiResizeTarget,
			&iiClock,
		)
		if err != nil {
			return nil, err
		}

		acc.Timestamp = accLoginTimestamp.Int64
		acc.Identicon = accIdenticon.String
		acc.ColorID = accColorID.Int64
		if len(accColorHash.String) != 0 {
			err = json.Unmarshal([]byte(accColorHash.String), &acc.ColorHash)
			if err != nil {
				return nil, err
			}
		}

		ii.KeyUID = acc.KeyUID
		ii.Name = iiName.String
		ii.Width = int(iiWidth.Int64)
		ii.Height = int(iiHeight.Int64)
		ii.FileSize = int(iiFileSize.Int64)
		ii.ResizeTarget = int(iiResizeTarget.Int64)
		ii.Clock = uint64(iiClock.Int64)

		if ii.Name == "" && len(ii.Payload) == 0 && ii.Width == 0 && ii.Height == 0 && ii.FileSize == 0 && ii.ResizeTarget == 0 {
			ii = nil
		}

		// Last index
		li := len(rst) - 1

		// Don't process nil identity images
		if ii != nil {
			// attach the identity image to a previously created account if present, check keyUID matches
			if len(rst) > 0 && rst[li].KeyUID == acc.KeyUID {
				rst[li].Images = append(rst[li].Images, *ii)
				// else attach the identity image to the newly created account
			} else {
				acc.Images = append(acc.Images, *ii)
			}
		}

		// Append newly created account only if this is the first loop or the keyUID doesn't match
		if len(rst) == 0 || rst[li].KeyUID != acc.KeyUID {
			rst = append(rst, acc)
		}
	}

	return rst, nil
}

func (db *Database) GetAccount(keyUID string) (*Account, error) {
	rows, err := db.db.Query("SELECT  a.name, a.loginTimestamp, a.identicon, a.colorHash, a.colorId, a.customizationColor, a.customizationColorClock, a.keycardPairing, a.keyUid, a.kdfIterations, ii.key_uid, ii.name, ii.image_payload, ii.width, ii.height, ii.file_size, ii.resize_target, ii.clock FROM accounts AS a LEFT JOIN identity_images AS ii ON ii.key_uid = a.keyUid WHERE a.keyUid = ? ORDER BY loginTimestamp DESC", keyUID)
	if err != nil {
		return nil, err
	}
	defer func() {
		errClose := rows.Close()
		err = valueOr(err, errClose)
	}()

	acc := new(Account)

	for rows.Next() {
		accLoginTimestamp := sql.NullInt64{}
		accIdenticon := sql.NullString{}
		accColorHash := sql.NullString{}
		accColorID := sql.NullInt64{}
		ii := &images.IdentityImage{}
		iiKeyUID := sql.NullString{}
		iiName := sql.NullString{}
		iiWidth := sql.NullInt64{}
		iiHeight := sql.NullInt64{}
		iiFileSize := sql.NullInt64{}
		iiResizeTarget := sql.NullInt64{}
		iiClock := sql.NullInt64{}

		err = rows.Scan(
			&acc.Name,
			&accLoginTimestamp,
			&accIdenticon,
			&accColorHash,
			&accColorID,
			&acc.CustomizationColor,
			&acc.CustomizationColorClock,
			&acc.KeycardPairing,
			&acc.KeyUID,
			&acc.KDFIterations,
			&iiKeyUID,
			&iiName,
			&ii.Payload,
			&iiWidth,
			&iiHeight,
			&iiFileSize,
			&iiResizeTarget,
			&iiClock,
		)
		if err != nil {
			return nil, err
		}

		acc.Timestamp = accLoginTimestamp.Int64
		acc.Identicon = accIdenticon.String
		acc.ColorID = accColorID.Int64
		if len(accColorHash.String) != 0 {
			err = json.Unmarshal([]byte(accColorHash.String), &acc.ColorHash)
			if err != nil {
				return nil, err
			}
		}

		ii.KeyUID = iiKeyUID.String
		ii.Name = iiName.String
		ii.Width = int(iiWidth.Int64)
		ii.Height = int(iiHeight.Int64)
		ii.FileSize = int(iiFileSize.Int64)
		ii.ResizeTarget = int(iiResizeTarget.Int64)
		ii.Clock = uint64(iiClock.Int64)

		// Don't process empty identity images
		if !ii.IsEmpty() {
			acc.Images = append(acc.Images, *ii)
		}
	}

	return acc, nil
}

func (db *Database) SaveAccount(account Account) error {
	colorHash, err := json.Marshal(account.ColorHash)
	if err != nil {
		return err
	}

	if account.KDFIterations <= 0 {
		account.KDFIterations = dbsetup.ReducedKDFIterationsNumber
	}

	_, err = db.db.Exec("INSERT OR REPLACE INTO accounts (name, identicon, colorHash, colorId, customizationColor, customizationColorClock, keycardPairing, keyUid, kdfIterations) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", account.Name, account.Identicon, colorHash, account.ColorID, account.CustomizationColor, account.CustomizationColorClock, account.KeycardPairing, account.KeyUID, account.KDFIterations)
	if err != nil {
		return err
	}

	if account.Images == nil {
		return nil
	}

	return db.StoreIdentityImages(account.KeyUID, account.Images, false)
}

func (db *Database) UpdateDisplayName(keyUID string, displayName string) error {
	_, err := db.db.Exec("UPDATE accounts SET name = ? WHERE keyUid = ?", displayName, keyUID)
	return err
}

func (db *Database) UpdateAccount(account Account) error {
	colorHash, err := json.Marshal(account.ColorHash)
	if err != nil {
		return err
	}

	if account.KDFIterations <= 0 {
		account.KDFIterations = dbsetup.ReducedKDFIterationsNumber
	}

	_, err = db.db.Exec("UPDATE accounts SET name = ?, identicon = ?, colorHash = ?, colorId = ?, customizationColor = ?, customizationColorClock = ?, keycardPairing = ?, kdfIterations = ? WHERE keyUid = ?", account.Name, account.Identicon, colorHash, account.ColorID, account.CustomizationColor, account.CustomizationColorClock, account.KeycardPairing, account.KDFIterations, account.KeyUID)
	return err
}

func (db *Database) UpdateAccountKeycardPairing(keyUID string, keycardPairing string) error {
	_, err := db.db.Exec("UPDATE accounts SET keycardPairing = ? WHERE keyUid = ?", keycardPairing, keyUID)
	return err
}

func (db *Database) UpdateAccountTimestamp(keyUID string, loginTimestamp int64) error {
	_, err := db.db.Exec("UPDATE accounts SET loginTimestamp = ? WHERE keyUid = ?", loginTimestamp, keyUID)
	return err
}

func (db *Database) UpdateAccountCustomizationColor(keyUID string, color string, clock uint64) (int64, error) {
	result, err := db.db.Exec("UPDATE accounts SET customizationColor = ?, customizationColorClock = ? WHERE keyUid = ? AND customizationColorClock < ?", color, clock, keyUID, clock)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *Database) DeleteAccount(keyUID string) error {
	_, err := db.db.Exec("DELETE FROM accounts WHERE keyUid = ?", keyUID)
	return err
}

// Account images
func (db *Database) GetIdentityImages(keyUID string) (iis []*images.IdentityImage, err error) {
	rows, err := db.db.Query(`SELECT key_uid, name, image_payload, width, height, file_size, resize_target, clock FROM identity_images WHERE key_uid = ?`, keyUID)
	if err != nil {
		return nil, err
	}
	defer func() {
		errClose := rows.Close()
		err = valueOr(err, errClose)
	}()

	for rows.Next() {
		ii := &images.IdentityImage{}
		err = rows.Scan(&ii.KeyUID, &ii.Name, &ii.Payload, &ii.Width, &ii.Height, &ii.FileSize, &ii.ResizeTarget, &ii.Clock)
		if err != nil {
			return nil, err
		}

		iis = append(iis, ii)
	}

	return iis, nil
}

func (db *Database) GetIdentityImage(keyUID, it string) (*images.IdentityImage, error) {
	var ii images.IdentityImage
	err := db.db.QueryRow("SELECT key_uid, name, image_payload, width, height, file_size, resize_target, clock FROM identity_images WHERE key_uid = ? AND name = ?", keyUID, it).Scan(&ii.KeyUID, &ii.Name, &ii.Payload, &ii.Width, &ii.Height, &ii.FileSize, &ii.ResizeTarget, &ii.Clock)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &ii, nil
}

func (db *Database) StoreIdentityImages(keyUID string, iis []images.IdentityImage, publish bool) (err error) {
	// Because SQL INSERTs are triggered in a loop use a tx to ensure a single call to the DB.
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}

		errRollback := tx.Rollback()
		err = valueOr(err, errRollback)
	}()

	for i, ii := range iis {
		if ii.IsEmpty() {
			continue
		}
		iis[i].KeyUID = keyUID
		_, err := tx.Exec(
			"INSERT INTO identity_images (key_uid, name, image_payload, width, height, file_size, resize_target, clock) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			keyUID,
			ii.Name,
			ii.Payload,
			ii.Width,
			ii.Height,
			ii.FileSize,
			ii.ResizeTarget,
			ii.Clock,
		)
		if err != nil {
			return err
		}
	}

	db.publishOnIdentityImageSubscriptions(&IdentityImageSubscriptionChange{
		PublishExpected: publish,
	})

	return nil
}

func (db *Database) SubscribeToIdentityImageChanges() chan *IdentityImageSubscriptionChange {
	s := make(chan *IdentityImageSubscriptionChange, 100)
	db.identityImageSubscriptions = append(db.identityImageSubscriptions, s)
	return s
}

func (db *Database) publishOnIdentityImageSubscriptions(change *IdentityImageSubscriptionChange) {
	// Publish on channels, drop if buffer is full
	for _, s := range db.identityImageSubscriptions {
		select {
		case s <- change:
		default:
			log.Warn("subscription channel full, dropping message")
		}
	}
}

func (db *Database) DeleteIdentityImage(keyUID string) error {
	_, err := db.db.Exec(`DELETE FROM identity_images WHERE key_uid = ?`, keyUID)

	if err != nil {
		return err
	}

	db.publishOnIdentityImageSubscriptions(&IdentityImageSubscriptionChange{
		PublishExpected: true,
	})

	return err
}

func valueOr(value error, or error) error {
	if value != nil {
		return value
	}
	return or
}
