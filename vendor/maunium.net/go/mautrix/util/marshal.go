package util

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// MarshalAndDeleteEmpty marshals a JSON object, then uses gjson to delete empty objects at the given gjson paths.
//
// This can be used as a convenient way to create a marshaler that omits empty non-pointer structs.
// See mautrix.RespSync for example.
func MarshalAndDeleteEmpty(marshalable interface{}, paths []string) ([]byte, error) {
	data, err := json.Marshal(marshalable)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		res := gjson.GetBytes(data, path)
		if res.IsObject() && len(res.Raw) == 2 {
			data, err = sjson.DeleteBytes(data, path)
			if err != nil {
				return nil, fmt.Errorf("failed to delete empty %s: %w", path, err)
			}
		}
	}
	return data, nil
}
