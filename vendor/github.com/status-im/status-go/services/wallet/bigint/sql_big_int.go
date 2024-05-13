package bigint

import (
	"database/sql/driver"
	"errors"
	"math/big"
)

// SQLBigInt type for storing uint256 in the databse.
// FIXME(dshulyak) SQL big int is max 64 bits. Maybe store as bytes in big endian and hope
// that lexographical sorting will work.
type SQLBigInt big.Int

// Scan implements interface.
func (i *SQLBigInt) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	val, ok := value.(int64)
	if !ok {
		return errors.New("not an integer")
	}
	(*big.Int)(i).SetInt64(val)
	return nil
}

// Value implements interface.
func (i *SQLBigInt) Value() (driver.Value, error) {
	val := (*big.Int)(i)
	if val == nil {
		return nil, nil
	}
	if !val.IsInt64() {
		return nil, errors.New("not an int64")
	}
	return (*big.Int)(i).Int64(), nil
}

// SQLBigIntBytes type for storing big.Int as BLOB in the databse.
type SQLBigIntBytes big.Int

func (i *SQLBigIntBytes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	val, ok := value.([]byte)
	if !ok {
		return errors.New("not an integer")
	}
	(*big.Int)(i).SetBytes(val)
	return nil
}

func (i *SQLBigIntBytes) Value() (driver.Value, error) {
	val := (*big.Int)(i)
	if val == nil {
		return nil, nil
	}
	return (*big.Int)(i).Bytes(), nil
}

type NilableSQLBigInt struct {
	big.Int
	isNil bool
}

func (i *NilableSQLBigInt) IsNil() bool {
	return i.isNil
}

func (i *NilableSQLBigInt) SetNil() {
	i.isNil = true
}

// Scan implements interface.
func (i *NilableSQLBigInt) Scan(value interface{}) error {
	if value == nil {
		i.SetNil()
		return nil
	}
	val, ok := value.(int64)
	if !ok {
		return errors.New("not an integer")
	}

	i.SetInt64(val)
	return nil
}

// Not implemented, used only for scanning
func (i *NilableSQLBigInt) Value() (driver.Value, error) {
	return nil, errors.New("NilableSQLBigInt.Value is not implemented")
}
