package typeddata

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

const (
	eip712Domain = "EIP712Domain"
	chainIDKey   = "chainId"
)

// Types define fields for each composite type.
type Types map[string][]Field

// Field stores name and solidity type of the field.
type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Validate checks that both name and type are not empty.
func (f Field) Validate() error {
	if len(f.Name) == 0 {
		return errors.New("`name` is required")
	}
	if len(f.Type) == 0 {
		return errors.New("`type` is required")
	}
	return nil
}

// TypedData defines typed data according to eip-712.
type TypedData struct {
	Types       Types                      `json:"types"`
	PrimaryType string                     `json:"primaryType"`
	Domain      map[string]json.RawMessage `json:"domain"`
	Message     map[string]json.RawMessage `json:"message"`
}

// Validate that required fields are defined.
// This method doesn't check if dependencies of the main type are defined, it will be validated
// when type string is computed.
func (t TypedData) Validate() error {
	if _, exist := t.Types[eip712Domain]; !exist {
		return fmt.Errorf("`%s` must be in `types`", eip712Domain)
	}
	if t.PrimaryType == "" {
		return errors.New("`primaryType` is required")
	}
	if _, exist := t.Types[t.PrimaryType]; !exist {
		return fmt.Errorf("primary type `%s` not defined in types", t.PrimaryType)
	}
	if t.Domain == nil {
		return errors.New("`domain` is required")
	}
	if t.Message == nil {
		return errors.New("`message` is required")
	}
	for typ := range t.Types {
		fields := t.Types[typ]
		for i := range fields {
			if err := fields[i].Validate(); err != nil {
				return fmt.Errorf("field %d from type `%s` is invalid: %v", i, typ, err)
			}
		}
	}
	return nil
}

// ValidateChainID accept chain as big integer and verifies if typed data belongs to the same chain.
func (t TypedData) ValidateChainID(chain *big.Int) error {
	if _, exist := t.Domain[chainIDKey]; !exist {
		return fmt.Errorf("domain misses chain key %s", chainIDKey)
	}
	var chainID int64
	if err := json.Unmarshal(t.Domain[chainIDKey], &chainID); err != nil {
		var chainIDString string
		if err = json.Unmarshal(t.Domain[chainIDKey], &chainIDString); err != nil {
			return err
		}
		if chainID, err = strconv.ParseInt(chainIDString, 0, 64); err != nil {
			return err
		}
	}
	if chainID != chain.Int64() {
		return fmt.Errorf("chainId %d doesn't match selected chain %s", chainID, chain)
	}
	return nil
}
