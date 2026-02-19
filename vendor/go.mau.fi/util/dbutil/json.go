package dbutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"unsafe"
)

// JSON is a utility type for using arbitrary JSON data as values in database Exec and Scan calls.
type JSON struct {
	Data any
}

func (j JSON) Scan(i any) error {
	switch value := i.(type) {
	case nil:
		return nil
	case string:
		return json.Unmarshal([]byte(value), j.Data)
	case []byte:
		return json.Unmarshal(value, j.Data)
	default:
		return fmt.Errorf("invalid type %T for dbutil.JSON.Scan", i)
	}
}

func (j JSON) Value() (driver.Value, error) {
	if j.Data == nil {
		return nil, nil
	}
	v, err := json.Marshal(j.Data)
	return unsafe.String(unsafe.SliceData(v), len(v)), err
}

// JSONPtr is a convenience function for wrapping a pointer to a value in the JSON utility, but removing typed nils
// (i.e. preventing nils from turning into the string "null" in the database).
func JSONPtr[T any](val *T) JSON {
	return JSON{Data: UntypedNil(val)}
}

func UntypedNil[T any](val *T) any {
	if val == nil {
		return nil
	}
	return val
}
