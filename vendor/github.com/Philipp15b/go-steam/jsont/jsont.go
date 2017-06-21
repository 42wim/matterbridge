// Includes helper types for working with JSON data
package jsont

import (
	"encoding/json"
)

// A boolean value that can be unmarshaled from a number in JSON.
type UintBool bool

func (u *UintBool) UnmarshalJSON(data []byte) error {
	var n uint
	err := json.Unmarshal(data, &n)
	if err != nil {
		return err
	}
	*u = n != 0
	return nil
}
