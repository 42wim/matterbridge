package sqlite

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
)

// JSONBlob type for marshaling/unmarshaling inner type to json.
type JSONBlob struct {
	Data  interface{}
	Valid bool
}

// Scan implements interface.
func (blob *JSONBlob) Scan(value interface{}) error {
	dataVal := reflect.ValueOf(blob.Data)
	blob.Valid = false
	if value == nil || dataVal.Kind() == reflect.Ptr && dataVal.IsNil() {
		return nil
	}

	var bytes []byte
	ok := true
	switch v := value.(type) {
	case []byte:
		bytes, ok = value.([]byte)
	case string:
		bytes = []byte(v)
	default:
		ok = false
	}
	if !ok {
		return errors.New("not a byte slice or string")
	}
	if len(bytes) == 0 {
		return nil
	}
	err := json.Unmarshal(bytes, blob.Data)
	blob.Valid = err == nil
	return err
}

// Value implements interface.
func (blob *JSONBlob) Value() (driver.Value, error) {
	dataVal := reflect.ValueOf(blob.Data)
	if (blob.Data == nil) || (dataVal.Kind() == reflect.Ptr && dataVal.IsNil()) {
		return nil, nil
	}

	switch dataVal.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		if dataVal.Len() == 0 {
			return nil, nil
		}
	}

	return json.Marshal(blob.Data)
}

func BigIntToClampedInt64(val *big.Int) *int64 {
	if val == nil {
		return nil
	}
	var v int64
	if val.IsInt64() {
		v = val.Int64()
	} else {
		v = math.MaxInt64
	}
	return &v
}

// BigIntToPadded128BitsStr converts a big.Int to a string, padding it with 0 to account for 128 bits size
// Returns nil if input val is nil
// This should work to sort and compare big.Ints values in SQLite
func BigIntToPadded128BitsStr(val *big.Int) *string {
	if val == nil {
		return nil
	}
	hexStr := val.Text(16)
	res := new(string)
	*res = fmt.Sprintf("%032s", hexStr)
	return res
}

func Int64ToPadded128BitsStr(val int64) *string {
	res := fmt.Sprintf("%032x", val)
	return &res
}
